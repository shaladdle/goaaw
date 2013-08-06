package rpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"strings"

	anet "aaw/net"
)

const ErrNil = StrError("")

type StrError string

func (s StrError) IsNil() bool {
	return s == ""
}

func (s StrError) Error() string {
	return string(s)
}

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

func NewPipeCliSrv(regName string, o interface{}) (*Client, *Server, error) {
	pnet := anet.NewPipeNet()

	srv := NewServer()
	srv.Register(regName, o)
	go srv.Accept(pnet)

	cli, err := NewClient(pnet)
	if err != nil {
		srv.Close()
		return nil, nil, err
	}

	return cli, srv, nil
}

type Server struct {
	coder   Coder
	types   map[string]reflect.Value // map of registered 'objects'
	methods map[string]method        // map of registered methods
	rpc     *rpc.Server
	closing chan bool
}

func NewServer() *Server {
	return NewServerWithCoder(defaultCoder)
}

func NewServerWithCoder(coder Coder) *Server {
	return &Server{
		coder:   coder,
		types:   make(map[string]reflect.Value),
		methods: make(map[string]method),
		rpc:     rpc.NewServer(),
		closing: make(chan bool),
	}
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
//  func RPCRead_methodNameHere(t1, t2, t3 ... , tn) (io.Reader, rt1, rt2 ... rtn)
//  func RPCWrite_methodNameHere(t1, t2, t3 ... , tn) (io.Writer, rt1, rt2 ... rtn)
//
//TODO: Support io.ReadCloser instead of io.Reader
func (s *Server) Register(name string, rcvr interface{}) error {
	const (
		norm_prefix  = "RPCNorm_"
		read_prefix  = "RPCRead_"
		write_prefix = "RPCWrite_"
	)

	refl := reflect.ValueOf(rcvr)
	typ := reflect.TypeOf(rcvr)
	s.types[name] = refl
	for i := 0; i < typ.NumMethod(); i++ {
		var (
			methodName string
			mClass     rpcClass
		)

		getName := func(i int, prefix string) string {
			return strings.SplitN(typ.Method(i).Name, prefix, 2)[1]
		}

		switch {
		case strings.HasPrefix(typ.Method(i).Name, norm_prefix):
			methodName = getName(i, norm_prefix)
			mClass = rpcNorm
		case strings.HasPrefix(typ.Method(i).Name, read_prefix):
			methodName = getName(i, read_prefix)
			mClass = rpcRead
		case strings.HasPrefix(typ.Method(i).Name, write_prefix):
			methodName = getName(i, write_prefix)
			mClass = rpcWrite
		default:
			continue
		}

		s.methods[name+"."+methodName] = method{refl.Method(i), mClass}
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
	accept := func(conns chan net.Conn) {
		conn, err := lis.Accept()
		if err != nil {
			close(conns)
			return
		}

		conns <- conn
	}

	for {
		conns := make(chan net.Conn)
		go accept(conns)

		var conn net.Conn

		select {
		case <-s.closing:
			lis.Close()
			s.closing <- true
			return
		case conn = <-conns:
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
	var methodName string
	if err := s.coder.Decode(conn, &methodName); err != nil {
		panic(err)
	}

	info, ok := s.methods[methodName]
	if !ok {
		panic(fmt.Sprintf("couldn't find %v in method index", methodName))
	}

	var args []interface{}
	if err := s.coder.Decode(conn, &args); err != nil {
		panic(err)
	}

	reflArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflArgs[i] = reflect.ValueOf(arg)
	}
	outs := info.method.Call(reflArgs)
	var sendOuts []reflect.Value

	switch info.class {
	case rpcNorm:
		sendOuts = outs
	default:
		sendOuts = outs[1:]
	}

	for _, out := range sendOuts {
		if err := s.coder.EncodeValue(conn, out); err != nil {
			log.Println(err)
			return
		}
	}

	switch info.class {
	case rpcRead:
		defer conn.Close()
		if outs[0].Interface() == nil {
			return
		}

		if _, err := io.Copy(conn, outs[0].Interface().(io.Reader)); err != nil {
			log.Println(err)
			return
		}
	case rpcWrite:
		if outs[0].Interface() == nil {
			return
		}

		w := outs[0].Interface().(io.WriteCloser)
		if _, err := io.Copy(w, conn); err != nil {
			log.Println(err)
			return
		}
		w.Close()
	}
}

func (s *Server) Close() {
	s.closing <- true
	<-s.closing
}
