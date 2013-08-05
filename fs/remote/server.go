package remote

import (
	"io"
    "net"

	"aaw/fs/std"
	anet "aaw/net"
	"aaw/rpc"
    "aaw/fs/util"
)

type Server struct {
	stdfs  std.FileSystem
	rpcSrv *rpc.Server
}

func NewPipeCliSrv(root string) (*Client, *Server, error) {
    pnet := anet.NewPipeNet()

    srv, err := NewServer(root, pnet)
    if err != nil {
        return nil, nil, err
    }

    cli, err := NewClient(pnet)
    if err != nil {
        srv.Close()
        return nil, nil, err
    }

    return cli, srv, nil
}

func NewServer(root string, l net.Listener) (*Server, error) {
	srv := &Server{
		stdfs:  std.New(root),
		rpcSrv: rpc.NewServer(),
	}

	srv.rpcSrv.Register("RemoteFS", srv)
	go srv.rpcSrv.Accept(l)

	return srv, nil
}

func NewTCPServer(root, hostport string) (*Server, error) {
	srv := &Server{
		stdfs:  std.New(root),
		rpcSrv: rpc.NewServer(),
	}

	srv.rpcSrv.Register("RemoteFS", srv)
	srv.rpcSrv.TCPListen(hostport)

	return srv, nil
}

func (s *Server) RPCWrite_Create(fpath string) (io.WriteCloser, rpc.StrError) {
	f, err := s.stdfs.Create(fpath)
	if err != nil {
		return nil, rpc.StrError(err.Error())
	}

	return f, rpc.ErrNil
}

func (s *Server) RPCRead_Open(fpath string) (io.Reader, rpc.StrError) {
	f, err := s.stdfs.Open(fpath)
	if err != nil {
		return nil, rpc.StrError(err.Error())
	}

	return f, rpc.ErrNil
}

func (s *Server) RPCNorm_Stat(fpath string) (util.FileInfo, rpc.StrError) {
	info, err := s.stdfs.Stat(fpath)
	if err != nil {
		return util.FileInfo{}, rpc.StrError(err.Error())
	}

	return util.FromOSInfo(info), rpc.ErrNil
}

func (s *Server) RPCNorm_Mkdir(fpath string) rpc.StrError {
	if err := s.stdfs.Mkdir(fpath); err != nil {
		return rpc.StrError(err.Error())
	}

	return rpc.ErrNil
}

func (s *Server) RPCNorm_Remove(fpath string) rpc.StrError {
	if err := s.stdfs.Remove(fpath); err != nil {
		return rpc.StrError(err.Error())
	}

	return rpc.ErrNil
}

func (s *Server) RPCNorm_GetFiles(fpath string) ([]util.FileInfo, rpc.StrError) {
	infos, err := s.stdfs.GetFiles(fpath)
	if err != nil {
		return nil, rpc.StrError(err.Error())
	}

    ret := make([]util.FileInfo, len(infos))
    for i, info := range infos {
        ret[i] = util.FromOSInfo(info)
    }

	return ret, rpc.ErrNil
}

func (s *Server) Close() {
	s.rpcSrv.Close()
}
