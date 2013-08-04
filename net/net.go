package net

import (
	"fmt"
	"net"
)

// Dialer describes objects capable of dialing based on pre-set parameters.
// This can be useful when the arguments for the dial are non-standard. One
// example is a TLS connection where configuration is required for the dial.
type Dialer interface {
	Dial() (net.Conn, error)
}

type TcpDialer string

func (d TcpDialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", string(d))
}

type Addr struct {
	network string
	str     string
}

func (a Addr) Network() string {
	return a.network
}

func (a Addr) String() string {
	return a.str
}

// TODO: Make PipeNet close properly. Having it there makes it satisfy the
// listener interface.
type PipeNet struct {
	conns chan net.Conn
}

func NewPipeNet() *PipeNet {
	return &PipeNet{
		make(chan net.Conn),
	}
}

func (p *PipeNet) Accept() (net.Conn, error) {
	return <-p.conns, nil
}

func (p *PipeNet) Close() error {
	return nil
}

func (p *PipeNet) Addr() net.Addr {
	return Addr{"pipe", "pipe"}
}

func (p *PipeNet) Dial() (net.Conn, error) {
	cli, srv := net.Pipe()

	p.conns <- srv

	return cli, nil
}

type chanListener chan net.Conn

func (cl chanListener) Accept() (net.Conn, error) {
	if conn, ok := <-cl; !ok {
		return nil, fmt.Errorf("accept on closed listener")
	} else {
		return conn, nil
	}
}

// TODO: This will currently cause a panic. Figure out a way to make it work
// even when someone is sending. Probably a director will solve it, since I
// can close outstanding accepts from the director.
func (cl chanListener) Close() error {
	close(cl)
	return nil
}

func (cl chanListener) Addr() net.Addr {
	return Addr{"rpc-chan-net", "rpc-chan-listener"}
}
