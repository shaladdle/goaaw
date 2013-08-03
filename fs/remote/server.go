package remote

type Server struct {
}

func NewServer(hostport string) (*Server, error) {
    rpcSrv := rpc.NewServer()
    rpcSrv.Register(srv)

    l, err := net.Listen("tcp", hostport)
    if err != nil {
        return nil, err
    }

    go rpcSrv.Accept(l)

    srv := &Server{}

    return srv, nil
}

func (s *Server) Open
