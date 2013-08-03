package server

import (
	"aaw/sync/remote/common"
	"net"
	"os"
	"path"
)

type Server struct {
    lpath    string
	hostport string
}

func New(hostport, lpath string) (*Server, error) {
	ret := &Server{
        lpath: lpath,
        hostport: hostport,
	}

	var err error

	l, err = net.Listen("tcp", hostport)
	if err != nil {
		return nil, err
	}

	go ret.director()

	return ret, nil
}

func (rs *Server) director() {
    handleNewClient := func(conn net.Conn) {
        defer conn.Close()

        msg, err := common.ReadMsg(conn)
        if err != nil {
            panic(err)
        }

        switch msg := msg.(type) {
        case common.GetMsg:
            rs.handleGet(conn, msg)
        case common.PutMsg:
            rs.handlePut(conn, msg)
        case common.DelMsg:
            rs.handleDel(msg)
        case common.GetIdxMsg:
            rs.handleGetIndex(conn, msg)
        default:
            panic("unrecognized message type")
        }
    }

	for {
        conn := <-rs.newClient

        go handleNewClient(conn)
	}
}

type getErrMsg struct {
	cli net.Conn
	err error
}

func (rs *Server) getFileName(id string) string {
	return path.Join(rs.lpath, id)
}

func (rs *Server) handleGet(conn net.Conn, msg common.GetMsg) {
    err := common.SendFile(conn, rs.getFileName(msg.Id))
    if err != nil {
        panic(err)
    }
}

func (rs *Server) handlePut(conn net.Conn, msg common.PutMsg) {
    err := common.RecvFile(conn, rs.getFileName(msg.Id))
    if err != nil {
        panic(err)
    }
}

func (rs *Server) handleDel(msg common.DelMsg) {
    err := os.Remove(rs.getFileName(msg.Id))
	if err != nil {
		panic(err)
	}
}

func (rs *Server) handleGetIndex(conn net.Conn, msg common.GetIdxMsg) {
    err := common.WriteIdx(conn, map[string]common.FileInfo{})
	if err != nil {
		panic(err)
	}
}
