package rpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"strings"
)

const (
	tagStreamRPC = iota
	tagNormRPC
	tagHandshake
)

type Coder interface {
	Encode(w io.Writer, src interface{}) error
	Decode(r io.Reader, dst interface{}) error
}

type jsonCoder struct { }

func (jsonCoder) Encode(w io.Writer, dst interface{}) error {
	b, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, int64(len(b)))
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (jsonCoder) Decode(r io.Reader, dst interface{}) error {
	var length int64
	err := binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return err
	}

	b := make([]byte, length)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, dst)
	if err != nil {
		return err
	}

	return nil
}

type method struct {
    method reflect.Value
    writer bool
}

type Server struct {
	coder   Coder
	types   map[string]reflect.Value // map of registered 'objects'
	methods map[string]method // map of registered methods
	start   chan bool
	conns   chan net.Conn
	rpc     *rpc.Server
}

func NewServer() *Server {
	return NewServerWithCoder(jsonCoder{})
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
// The following are the different method signatures.
//
// Standard RPC methods:
//  func(args t1, reply *t2) error
// see net/rpc docs for more info.
//
// Streaming write methods:
//  func(arg1 t1, arg2 t2, arg3 t3 ...) (io.Writer, error)
//
// Streaming read methods:
//  func(arg1 t1, arg2 t2, arg3 t3 ...) (io.Reader, error)
func (s *Server) Register(name string, rcvr interface{}) error {
	typ := reflect.TypeOf(rcvr)

	if _, ok := s.types[name]; ok {
		return fmt.Errorf("already registered %v", name)
	}

	if err := s.rpc.Register(rcvr); err != nil {
		return err
	}

	s.types[name] = reflect.ValueOf(rcvr)

    tmp := reflect.TypeOf(func() (io.Reader, io.Writer, error) { return nil, nil, nil })
    writerTyp := tmp.Out(0)
    readerTyp := tmp.Out(1)
    errorTyp := tmp.Out(2)

	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if m.Type.NumOut() != 2 {
            log.Printf("stream: method %v two return values %v", m.Name, m.Type.NumOut())
			continue
		}

		if m.Type.Out(1) != errorTyp {
            log.Printf("stream: method %v should have second return type of error, got %v", m.Name, m.Type.Out(1))
			continue
		}

        isReader := m.Type.Out(0) == readerTyp
        isWriter := m.Type.Out(0) == writerTyp
		if !isReader && !isWriter {
            log.Printf("stream: method %v should have either io.Reader or io.Writer as return type", m.Name)
			continue
		}

		s.methods[name+"."+m.Name] = method{m.Func, isWriter}
	}

	return nil
}

func (s *Server) Accept(lis net.Listener) {
	conns := make(chanListener)

	go s.rpc.Accept(conns)

	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		var tag byte
		err = s.coder.Decode(conn, &tag)
		if err != nil {
			panic(err)
		}

		switch tag {
		case tagHandshake:
            log.Println("handshake")
			go s.handshake(conn)
		case tagStreamRPC:
            log.Println("streaming")
			// Each streaming RPC call gets its own connection
			go s.handleStreamRPC(conn)
		case tagNormRPC:
            log.Println("normal")
			// Hand this off to the stdlib rpc library
			conns <- conn
		default:
			panic("Unrecognized message")
		}
	}
}

func (s *Server) handshake(conn net.Conn) {
	methods := make(map[string]bool)
	for k := range s.methods {
		methods[k] = true
	}

	err := s.coder.Encode(conn, methods)
	if err != nil {
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
	errInt := outs[1].Interface()
	if errInt != nil {
		err := errInt.(error)
		log.Printf("Error on stream rpc method call %v(%v): %v", mName, args, err)
		return
	}

	// TODO: Add support for writing RPCs too. This might just entail doing
	// a type switch on outs[0] and if it's a writer, copy the other direction

	// TODO: Add a byte to the protocol that specifies whether an error is
	// being returned or a stream is coming back. For now we just assume
	// there was no error so the client starts the receive
    if s.methods[mName].writer {
        o := outs[0].Interface().(io.Writer)
        log.Println("in writer case")
		_, err = io.Copy(o, conn)
		if err != nil {
			log.Println("Error executing copy:", err)
			return
		}
    } else {
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
