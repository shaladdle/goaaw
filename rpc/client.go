package rpc

import (
	"fmt"
	"io"
	"net"
	"reflect"

	anet "aaw/net"
)

type Client struct {
	coder    Coder
	d        anet.Dialer
	rpcIndex map[string]rpcClass
}

func NewClient(d anet.Dialer) (*Client, error) {
	return NewClientWithCoder(d, defaultCoder)
}

func NewClientWithCoder(d anet.Dialer, coder Coder) (*Client, error) {
	rpcIndex, err := clientDoHandshake(d, coder)
	if err != nil {
		return nil, err
	}

	ret := &Client{
		coder:    coder,
		d:        d,
		rpcIndex: rpcIndex,
	}

	return ret, nil
}

// Call actually goes over the network and performs the call. Args does double
// as both arguments and return values for normal RPCs. The return empty
// interface is either a io.ReadCloser, an io.Writer, or a nil, depending on
// the type of RPC.
func (c *Client) Call(methodName string, fnargs ...interface{}) error {
	if class, ok := c.rpcIndex[methodName]; !ok {
		return fmt.Errorf("could not find rpc %v", methodName)
	} else if class != rpcNorm {
		return fmt.Errorf("wrong rpc type, please use CallRead or CallWrite for streaming RPCs")
	}

	conn, err := c.call(methodName, fnargs)
	if err != nil {
		return err
	}

	return conn.Close()
}

func (c *Client) CallRead(methodName string, fnargs ...interface{}) (io.Reader, error) {
	if class, ok := c.rpcIndex[methodName]; !ok {
		return nil, fmt.Errorf("could not find rpc %v", methodName)
	} else if class != rpcRead {
		return nil, fmt.Errorf("wrong rpc type, is this a write or normal rpc?")
	}

	return c.call(methodName, fnargs)
}

func (c *Client) CallWrite(methodName string, fnargs ...interface{}) (io.WriteCloser, error) {
	if class, ok := c.rpcIndex[methodName]; !ok {
		return nil, fmt.Errorf("could not find rpc %v", methodName)
	} else if class != rpcWrite {
		return nil, fmt.Errorf("wrong rpc type, is this a read or normal rpc?")
	}

	return c.call(methodName, fnargs)
}

func (c *Client) call(methodName string, fnargs []interface{}) (net.Conn, error) {
	conn, err := c.d.Dial()
	if err != nil {
		return nil, err
	}

	// Indicate that this is an RPC connection.
	if err := c.coder.Encode(conn, tagRPC); err != nil {
		return nil, err
	}

	// Send method name.
	if err := c.coder.Encode(conn, methodName); err != nil {
		return nil, err
	}

	args := []interface{}{}
	rets := []interface{}{}
	for _, fnarg := range fnargs {
		r := reflect.TypeOf(fnarg)
		if r.Kind() == reflect.Ptr {
			rets = append(rets, fnarg)
		} else {
			args = append(args, fnarg)
		}
	}

	// Send arguments.
	if err := c.coder.Encode(conn, args); err != nil {
		return nil, err
	}

	// Get return values.
	var readRets []interface{}
	if err := c.coder.Decode(conn, &readRets); err != nil {
		return nil, err
	}

	if len(readRets) != len(rets) {
		return nil, fmt.Errorf("wrong number of returns decoded, expected %v, got %v", len(rets), len(readRets))
	}

	for i, read := range readRets {
		reflect.ValueOf(rets[i]).Elem().Set(reflect.ValueOf(read))
	}

	return conn, nil
}

func (c *Client) Close() error {
	return nil
}

func clientDoHandshake(d anet.Dialer, coder Coder) (map[string]rpcClass, error) {
	handshakeConn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	defer handshakeConn.Close()

	rpcIndex, err := getStreamRPCList(handshakeConn, coder)
	if err != nil {
		return nil, err
	}

	return rpcIndex, nil
}

func getStreamRPCList(conn net.Conn, coder Coder) (map[string]rpcClass, error) {
	err := coder.Encode(conn, tagHandshake)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]rpcClass)
	err = coder.Decode(conn, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
