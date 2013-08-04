package rpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"strings"
)

const (
	tagRPC = byte(iota)
	tagHandshake
)

type rpcClass byte

const (
	rpcNorm = rpcClass(iota)
	rpcWrite
	rpcRead
)

type method struct {
	method reflect.Value
	class  rpcClass
}

type Server struct {
	coder   Coder
	types   map[string]reflect.Value // map of registered 'objects'
	methods map[string]method        // map of registered methods
	start   chan bool
	conns   chan net.Conn
	rpc     *rpc.Server
}

func NewServer() *Server {
	return NewServerWithCoder(defaultCoder)
}

func NewServerWithCoder(coder Coder) *Server {
	return &Server{
		coder:   coder,
		types:   make(map[string]reflect.Value),
		methods: make(map[string]method),
		start:   make(chan bool),
		conns:   make(chan net.Conn),
		rpc:     rpc.NewServer(),
	}
}

func (s *Server) Close() error {
	return nil
}

// Register registers an object with this rpc server. It returns an error if a
// type is not suitable for use as an rpc server. Methods satisfying the
// streaming rpc signature will be registered as streaming functions, and
// methods satisfying the standard library's rpc function signature will
// be registered with an instance of that rpc server.
//
// Method signatures must fit one of the following forms. For normal rpcs that
// do not stream, any number of arguments and return values must be defined.
//  func RPCNorm_methodNameHere(t1, t2, t3 ... , tn) (rt1, rt2 ... rtn)
//  func RPCRead_methodNameHere(t1, t2, t3 ... , tn) (io.ReadCloser, rt1, rt2 ... rtn)
//  func RPCWrite_methodNameHere(t1, t2, t3 ... , tn) (io.Writer, rt1, rt2 ... rtn)
func (s *Server) Register(name string, rcvr interface{}) error {
	const (
		norm_prefix  = "RPCNorm_"
		read_prefix  = "RPCRead_"
		write_prefix = "RPCWrite_"
	)

	refl := reflect.ValueOf(rcvr)
	typ := reflect.TypeOf(rcvr)
	s.types[name] = refl
	for i := 0; i < refl.NumMethod(); i++ {
		var (
			mName  string
			mClass rpcClass
		)

		getName := func(prefix string) string {
			return strings.SplitN(typ.Method(i).Name, prefix, 2)[1]
		}

		switch {
		case strings.HasPrefix(typ.Method(i).Name, norm_prefix):
			mName = getName(norm_prefix)
			mClass = rpcNorm
		case strings.HasPrefix(typ.Method(i).Name, read_prefix):
			mName = getName(read_prefix)
			mClass = rpcRead
		case strings.HasPrefix(typ.Method(i).Name, write_prefix):
			mName = getName(write_prefix)
			mClass = rpcWrite
		default:
			continue
		}

		s.methods[name+"."+mName] = method{refl.Method(i), mClass}
	}

	return nil
}

func (s *Server) TCPListen(hostport string) error {
	l, err := net.Listen("tcp", hostport)
	if err != nil {
		return err
	}

	go s.Accept(l)

	return nil
}

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		var tag byte

		if err := s.coder.Decode(conn, &tag); err != nil {
			panic(err)
		}

		switch tag {
		case tagHandshake:
			go s.handshake(conn)
		case tagRPC:
			go s.handleRPC(conn)
		default:
			panic("Unrecognized message")
		}
	}
}

func (s *Server) handshake(conn net.Conn) {
	methods := make(map[string]rpcClass)
	for i, m := range s.methods {
		methods[i] = m.class
	}

	err := s.coder.Encode(conn, methods)
	if err != nil {
		panic(err)
	}
}

func (s *Server) handleRPC(conn net.Conn) {
	var (
		class      rpcClass
		methodName string
	)

	if err := s.coder.Decode(conn, &class); err != nil {
		panic(err)
	}

	if err := s.coder.Decode(conn, &methodName); err != nil {
		panic(err)
	}

	_, ok := s.methods[methodName]
	if !ok {
		panic(fmt.Sprintf("couldn't find %v in method index", methodName))
	}

	switch class {
	case rpcNorm:
		s.handleNormRPC(conn, methodName)
	default:
		panic(fmt.Sprintf("not doing that yet: %v", class))
	}
}

func (s *Server) handleNormRPC(conn net.Conn, methodName string) {
	info := s.methods[methodName]

	var args []interface{}
	if err := s.coder.Decode(conn, &args); err != nil {
		panic(err)
	}

	reflArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflArgs[i] = reflect.ValueOf(arg)
	}
	outs := info.method.Call(reflArgs)

	rets := make([]interface{}, len(outs))
	for i, out := range outs {
		rets[i] = out.Interface()
	}

	if err := s.coder.Encode(conn, rets); err != nil {
		panic(err)
	}
}

// For writing RPCs, we want to decode arguments and pass them as well as
// the conn to the local method so it can read.
func (s *Server) handleStreamRPC(conn net.Conn) {
	var mName string
	err := s.coder.Decode(conn, &mName)
	if err != nil {
		log.Println("Received malformed stream rpc, couldn't decode method name")
		conn.Close()
		return
	}

	// TODO: Add support for arguments of any type

	// TODO: Add support for any number of arguments. I think think an be done
	// by just iterating through all the arguments s.methods[mName] and
	// decoding to reflect.New'ed pointers for each type

	mth, ok := s.methods[mName]
	if !ok {
		log.Printf("Method %v not recognized", mName)
		conn.Close()
		return
	}

	mType := mth.method.Type()

	name := strings.SplitN(mName, ".", 2)[0]

	// Decode arguments
	args := make([]reflect.Value, mType.NumIn())
	args[0] = s.types[name]
	for i := 1; i < mType.NumIn(); i++ {
		val := reflect.New(mType.In(i))
		err = s.coder.Decode(conn, val.Interface())
		if err != nil {
			log.Println("couldn't decode argument %v of type %v", i, mType.In(i))
			return
		}

		args[i] = val.Elem()
	}

	// TODO: For a stream write, we are emulating a function with an io.Reader
	// argument. Pass conn in as the first argument, and the decoded arguments
	// as the rest.
	outs := s.methods[mName].method.Call(args)

	outInts := make([]interface{}, len(outs))
	for i, o := range outs {
		outInts[i] = o.Interface()
	}

	switch s.methods[mName].class {
	case rpcNorm:
		err = s.coder.Encode(conn, outInts)
		if err != nil {
			log.Println("Error writing return arguments to client:", err)
			return
		}
	case rpcWrite:
		o := outs[0].Interface().(io.Writer)
		log.Println("in writer case")
		_, err = io.Copy(o, conn)
		if err != nil {
			log.Println("Error executing copy:", err)
			return
		}
	case rpcRead:
		o := outs[0].Interface().(io.Reader)
		_, err = io.Copy(conn, o)
		if err != nil {
			log.Println("Error executing copy:", err)
			return
		}

		log.Println("in reader case")

		err = conn.Close()
		if err != nil {
			log.Println("Error closing connection after copy:", err)
			return
		}
	}
}
