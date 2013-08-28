package blkstore

import (
	"github.com/shaladdle/goaaw/fs/remote"
	anet "github.com/shaladdle/goaaw/net"
)

func NewRemoteStore(d anet.Dialer) (BlkStore, error) {
	cli, err := remote.NewClient(d)
	if err != nil {
		return nil, err
	}

	return &diskstore{cli}, nil
}
