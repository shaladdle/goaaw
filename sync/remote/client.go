package remote

import (
	"errors"
	"io"
    "log"
	"net"
	"os"

	"aaw/rpc"
)

type Index interface {
	Stat(key string) os.FileInfo
	Exists(key string) bool
}

type dialer struct {
	hostport string
}

func (d dialer) Dial() (net.Conn, error) {
	return net.Dial("tcp", d.hostport)
}

// Client exposes the server's functionality as though it were a local library
type Client struct {
	rpcCli *rpc.Client
}

func NewClient(hostport string) (*Client, error) {
	rpcCli, err := rpc.NewClient(dialer{hostport})
	if err != nil {
		return nil, err
	}

	return &Client{rpcCli}, nil
}

func (c *Client) Get(key, dstPath string) error {
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := c.rpcCli.CallRead("SyncServer.Open", key)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Put(key, srcPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := c.rpcCli.CallWrite("SyncServer.Create", key)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, f)
	if err != nil {
        log.Println("hi")
		return err
	}

	return nil
}

func (c *Client) Delete(key string) error {
	var resp Response
	err := c.rpcCli.Call("SyncServer.Remove", key, &resp)
	if err != nil {
		return err
	}

	if resp.Status != ST_OK {
		return errors.New(resp.ErrMsg)
	}

	return nil
}
