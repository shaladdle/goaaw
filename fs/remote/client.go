package remote

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

// Client satisfies the FileSystem interface defined in aaw/fs.
type Client struct {
}

func NewClient(hostport string) *Client {
    return &Client{
    }
}

func (fs *Client) Open(fpath string) (io.ReadCloser, error)     {}
func (fs *Client) Mkdir(dpath string) error                     {}
func (fs *Client) Create(fpath string) (io.WriteCloser, error)  {}
func (fs *Client) Stat(fpath string) (os.FileInfo, error)       {}
func (fs *Client) Remove(fpath string) error                    {}
func (fs *Client) GetFiles(fpath string) ([]os.FileInfo, error) {}
