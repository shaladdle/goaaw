package rpc

import (
	"io"
	"net"
	"net/rpc"
)

type Client struct {
	coder      Coder
	d          Dialer
	rpc        *rpc.Client
	streamRPCs map[string]bool
}

func NewClient(d Dialer) (*Client, error) {
	coder := jsonCoder{}

	rpcConn, err := d.Dial()
	if err != nil {
		return nil, err
	}

	err = coder.Encode(rpcConn, tagNormRPC)
	if err != nil {
		return nil, err
	}

	handshakeConn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	defer handshakeConn.Close()

	streamRPCs, err := getStreamRPCList(handshakeConn)
	if err != nil {
		return nil, err
	}

	ret := &Client{
		coder:      coder,
		d:          d,
		rpc:        rpc.NewClient(rpcConn),
		streamRPCs: streamRPCs,
	}

	return ret, nil
}

// Call actually goes over the network and performs the call
func (c *Client) Call(mName string, args interface{}, reply interface{}) error {
    return c.rpc.Call(mName, args, reply)
}

// CallWrite starts a streaming write RPC.
//
// Example:
//  f, err := c.CallWrite("File.Create", "myfile.txt")
//  // Use f like you called the standard library file functions
func (c *Client) CallWrite(mName string, args ...interface{}) (io.WriteCloser, error) {
	conn, err := c.callStreaming(mName, args)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// CallRead starts a streaming read RPC.
//
// Example:
//  r, err := c.CallWrite("File.Open", "myfile.txt")
//  if err != nil {
//      fmt.Println("There was an error opening myfile.txt")
//  }
//
//  for {
//      b := make([]byte, 1024)
//      n, err := r.Read(b)
//      if err != nil {
//          fmt.Println("There was an error opening myfile.txt")
//          break
//      }
//      fmt.Println(string(b[:n]))
//  }
func (c *Client) CallRead(mName string, args ...interface{}) (io.Reader, error) {
	conn, err := c.callStreaming(mName, args)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) callStreaming(mName string, args []interface{}) (net.Conn, error) {
	conn, err := c.d.Dial()
	if err != nil {
		return nil, err
	}

	err = c.coder.Encode(conn, tagStreamRPC)
	if err != nil {
		return nil, err
	}

	err = c.coder.Encode(conn, mName)
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		err := c.coder.Encode(conn, arg)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func getStreamRPCList(conn net.Conn) (map[string]bool, error) {
	coder := jsonCoder{}

	err := coder.Encode(conn, tagHandshake)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]bool)
	err = coder.Decode(conn, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
