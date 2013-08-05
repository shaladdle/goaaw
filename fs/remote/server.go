package remote

import (
	"encoding/gob"
	"io"
	"os"
	"time"
    "net"

	"aaw/fs/std"
	anet "aaw/net"
	"aaw/rpc"
)

func init() {
	gob.Register(fileInfo{})
	gob.Register([]fileInfo{})
}

type fileInfo struct {
	I_Name    string      // base name of the file
	I_Size    int64       // length in bytes for regular files; system-dependent for others
	I_Mode    os.FileMode // file mode bits
	I_ModTime time.Time   // modification time
	I_IsDir   bool        // abbreviation for Mode().IsDir()
}

func (info fileInfo) Name() string       { return info.I_Name }
func (info fileInfo) Size() int64        { return info.I_Size }
func (info fileInfo) Mode() os.FileMode  { return info.I_Mode }
func (info fileInfo) ModTime() time.Time { return info.I_ModTime }
func (info fileInfo) IsDir() bool        { return info.I_IsDir }
func (info fileInfo) Sys() interface{}   { return nil }

func fromOSFileInfo(info os.FileInfo) fileInfo {
	return fileInfo{
		I_Name:    info.Name(),
		I_Size:    info.Size(),
		I_Mode:    info.Mode(),
		I_ModTime: info.ModTime(),
		I_IsDir:   info.IsDir(),
	}
}

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

func (s *Server) RPCNorm_Stat(fpath string) (fileInfo, rpc.StrError) {
	info, err := s.stdfs.Stat(fpath)
	if err != nil {
		return fileInfo{}, rpc.StrError(err.Error())
	}

	return fromOSFileInfo(info), rpc.ErrNil
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

func (s *Server) RPCNorm_GetFiles(fpath string) ([]fileInfo, rpc.StrError) {
	infos, err := s.stdfs.GetFiles(fpath)
	if err != nil {
		return nil, rpc.StrError(err.Error())
	}

    ret := make([]fileInfo, len(infos))
    for i, info := range infos {
        ret[i] = fromOSFileInfo(info)
    }

	return ret, rpc.ErrNil
}

func (s *Server) Close() {
	s.rpcSrv.Close()
}
