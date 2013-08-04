package net

import (
	"net"
	"testing"
)

func TestChanListener(t *testing.T) {
	cl := make(chanListener)

	conn := &net.TCPConn{}
	go func() {
		cl <- conn
	}()

	rconn, err := cl.Accept()
	if err != nil {
		t.Fatal(err)
	}

	if rconn != conn {
		t.Fatal("conns don't match")
	}
}
