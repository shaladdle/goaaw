package blkstore

import (
	"aaw/fs/remote"
	anet "aaw/net"
)

func NewRemoteStore(d anet.Dialer) (BlkStore, error) {
	cli, err := remote.NewClient(d)
	if err != nil {
		return nil, err
	}

	return &diskstore{cli}, nil
}
