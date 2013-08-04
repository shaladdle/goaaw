package rpc

import (
	"encoding/gob"
	"fmt"
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
func (c *Client) Call(methodName string, fnargs ...interface{}) (interface{}, error) {
	conn, err := c.d.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	switch c.rpcIndex[methodName] {
	case rpcNorm:
		_, err := c.handleNormRPC(methodName, conn, fnargs)
		return nil, err
	case rpcRead:
	case rpcWrite:
	}

	return nil, fmt.Errorf("should not reach here")
}

func (c *Client) handleNormRPC(methodName string, conn net.Conn, fnargs []interface{}) (interface{}, error) {
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

	// Indicate that this is an RPC connection.
	if err := c.coder.Encode(conn, tagRPC); err != nil {
		return nil, err
	}

	// Indicate that we are doing a normal RPC, not a streaming one.
	if err := c.coder.Encode(conn, rpcNorm); err != nil {
		return nil, err
	}

	// Send method name.
	if err := c.coder.Encode(conn, methodName); err != nil {
		return nil, err
	}

	// Send arguments.
	if err := c.coder.Encode(conn, args); err != nil {
		return nil, err
	}

	gc := gob.NewDecoder(conn)
	// Get return values.
	var readRets []interface{}
	if err := gc.Decode(&readRets); err != nil {
		return nil, err
	}

	if len(readRets) != len(rets) {
		return nil, fmt.Errorf("wrong number of returns decoded, expected %v, got %v", len(rets), len(readRets))
	}

	for i, read := range readRets {
		reflRet := reflect.ValueOf(rets[i])
		reflRet.Elem().Set(reflect.ValueOf(read))
	}

	return nil, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) callStreaming(mName string, args []interface{}) (net.Conn, error) {
	conn, err := c.d.Dial()
	if err != nil {
		return nil, err
	}

	err = c.coder.Encode(conn, tagRPC)
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
