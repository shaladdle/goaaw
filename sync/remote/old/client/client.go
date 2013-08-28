package client

import (
	"github.com/shaladdle/goaaw/sync/remote/common"
	"net"
	"path"
)

type dGetMsg struct {
	id    string
	err   chan error
}

type dPutMsg struct {
	id  string
	err chan error
}

type dDelMsg struct {
	id  string
	err chan error
}

type dGetIndexMsg struct {
	reply chan map[string]common.FileInfo
	err   chan error
}

type client struct {
	msg      chan interface{}
	maxConns int
	raddr    string
	lpath    string
}

func New(raddr, lpath string) (common.Storage, error) {
	ret := &client{
		msg:      make(chan interface{}),
		raddr:    raddr,
		maxConns: 10,
		lpath:    lpath,
	}

	go ret.director()

	return ret, nil
}

func (cli *client) handlePut(msg dPutMsg) {
	conn, err := net.Dial("tcp", cli.raddr)
	if err != nil {
		msg.err <- err
		return
	}
    defer conn.Close()

	err = common.WritePutMsg(conn, msg.id)
	if err != nil {
		msg.err <- err
		return
	}

	msg.err <- common.SendFile(conn, path.Join(cli.lpath, msg.id))
}

func (cli *client) handleGet(msg dGetMsg) {
	conn, err := net.Dial("tcp", cli.raddr)
	if err != nil {
		msg.err <- err
		return
	}
    defer conn.Close()

	err = common.WriteGetMsg(conn, msg.id)
	if err != nil {
		msg.err <- err
		return
	}

	msg.err <- common.RecvFile(conn, path.Join(cli.lpath, msg.id))
}

func (cli *client) handleDel(msg dDelMsg) {
	conn, err := net.Dial("tcp", cli.raddr)
	if err != nil {
		msg.err <- err
		return
	}
    defer conn.Close()

	msg.err <- common.WriteDelMsg(conn, msg.id)
}

func (cli *client) handleGetIndex(msg dGetIndexMsg) {
	conn, err := net.Dial("tcp", cli.raddr)
	if err != nil {
		msg.err <- err
		return
	}
    defer conn.Close()

    err = common.WriteGetIdxMsg(conn)
	if err != nil {
		msg.err <- err
		return
	}

    ret, err := common.ReadIdx(conn)
	if err != nil {
		msg.err <- err
		return
	}

	msg.reply <- ret
}

func (cli *client) director() {
	for {
		msg := <-cli.msg

		switch msg := msg.(type) {
		case dGetMsg:
			cli.handleGet(msg)
		case dPutMsg:
			cli.handlePut(msg)
		case dDelMsg:
			cli.handleDel(msg)
		case dGetIndexMsg:
			cli.handleGetIndex(msg)
		}
	}
}
