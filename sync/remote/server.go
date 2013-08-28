package remote

import (
	"io"
	"net"
	"os"
	"path"

	"github.com/shaladdle/goaaw/rpc"
)

const (
	ST_OK  = 0
	ST_ERR = 1
)

type Response struct {
	Status byte
	ErrMsg string
}

type RpcServer struct {
	root  string
	fatal chan error
}

func newServer(root string) *RpcServer {
	return &RpcServer{
		root:  root,
		fatal: make(chan error),
	}
}

func (rs *RpcServer) getPath(key string) string {
	return path.Join(rs.root, key)
}

func (rs *RpcServer) Open(key string) (io.Reader, error) {
	return os.Open(rs.getPath(key))
}

func (rs *RpcServer) Create(key string) (io.Writer, error) {
	return os.Create(rs.getPath(key))
}

func (rs *RpcServer) Remove(key string, resp *Response) error {
	err := os.Remove(rs.getPath(key))
	if err != nil {
		*resp = Response{ST_ERR, err.Error()}
	}

	return nil
}

type Server struct {
	srv    *RpcServer
	rpcSrv *rpc.Server
}

func NewServer(hostport, root string) (*Server, error) {
	s := &Server{
		srv:    newServer(root),
		rpcSrv: rpc.NewServer(),
	}

	err := s.rpcSrv.Register("SyncServer", s.srv)
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", hostport)
	if err != nil {
		return nil, err
	}

	go s.rpcSrv.Accept(l)

	return s, nil
}
