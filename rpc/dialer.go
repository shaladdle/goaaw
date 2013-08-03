package rpc

import (
    "net"
)

// Dialer describes some interface that's capable of creating new connections
// to a server. The intended use for this is to allow users to provide dialers
// that use encrypted connections, without making the rpc library depend on any
// of the encryption details.
type Dialer interface {
	Dial() (net.Conn, error)
}

type TcpDialer string

func (d TcpDialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", string(d))
}
