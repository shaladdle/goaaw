package remote

import (
	"fmt"
	"io"
	"os"

	"github.com/shaladdle/goaaw/fs/util"
	anet "github.com/shaladdle/goaaw/net"
	"github.com/shaladdle/goaaw/rpc"
)

type closeWrapper struct {
	io.Reader
}

func (closeWrapper) Close() error {
	return nil
}

// Client satisfies the FileSystem interface defined in aaw/fs.
type Client struct {
	rpc *rpc.Client
}

func NewClient(d anet.Dialer) (*Client, error) {
	cli, err := rpc.NewClient(d)
	if err != nil {
		return nil, err
	}

	return &Client{cli}, nil
}

func NewTCPClient(hostport string) (*Client, error) {
	cli, err := rpc.NewClient(anet.TCPDialer(hostport))
	if err != nil {
		return nil, err
	}

	return &Client{cli}, nil
}

func (fs *Client) Create(fpath string) (io.WriteCloser, error) {
	var cErr rpc.StrError

	f, err := fs.rpc.CallWrite("RemoteFS.Create", fpath, &cErr)
	if err != nil {
		return nil, fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return nil, cErr
	}

	return f, nil
}

func (fs *Client) Open(fpath string) (io.ReadCloser, error) {
	var cErr rpc.StrError

	f, err := fs.rpc.CallRead("RemoteFS.Open", fpath, &cErr)
	if err != nil {
		return nil, fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return nil, cErr
	}

	return closeWrapper{f}, nil
}

func (fs *Client) Stat(fpath string) (os.FileInfo, error) {
	var (
		cErr rpc.StrError
		info util.FileInfo
	)

	err := fs.rpc.Call("RemoteFS.Stat", fpath, &info, &cErr)
	if err != nil {
		return nil, fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return nil, cErr
	}

	return info, nil
}

func (fs *Client) Mkdir(fpath string) error {
	var cErr rpc.StrError

	err := fs.rpc.Call("RemoteFS.Mkdir", fpath, &cErr)
	if err != nil {
		return fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return cErr
	}

	return nil
}

func (fs *Client) Remove(fpath string) error {
	var cErr rpc.StrError

	err := fs.rpc.Call("RemoteFS.Remove", fpath, &cErr)
	if err != nil {
		return fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return cErr
	}

	return nil
}

func (fs *Client) GetFiles(fpath string) ([]os.FileInfo, error) {
	var (
		cErr  rpc.StrError
		infos []util.FileInfo
	)

	err := fs.rpc.Call("RemoteFS.GetFiles", fpath, &infos, &cErr)
	if err != nil {
		return nil, fmt.Errorf("rpc error: %v", err)
	}

	if !cErr.IsNil() {
		return nil, cErr
	}

	ret := make([]os.FileInfo, len(infos))
	for i, info := range infos {
		ret[i] = info
	}

	return ret, nil
}

func (fs *Client) Close() {}
