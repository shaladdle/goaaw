package rpc

import (
	"fmt"
	"net"
)

type addr struct {
	network string
	str     string
}

func (a addr) Network() string {
	return a.network
}

func (a addr) String() string {
	return a.str
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
	return addr{"rpc-chan-net", "rpc-chan-listener"}
}
