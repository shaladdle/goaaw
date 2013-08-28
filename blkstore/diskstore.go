package blkstore

import (
	"fmt"
	"io/ioutil"

	"github.com/shaladdle/goaaw/fs"
	"github.com/shaladdle/goaaw/fs/std"
)

type diskstore struct {
	disk fs.FileSystem
}

func NewDiskStore(root string) BlkStore {
	return &diskstore{std.New(root)}
}

func (bs *diskstore) Get(key string) ([]byte, error) {
	f, err := bs.disk.Open(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func (bs *diskstore) Put(key string, blk []byte) error {
	f, err := bs.disk.Create(key)
	if err != nil {
		return err
	}
	defer f.Close()

	if n, err := f.Write(blk); err != nil {
		return err
	} else if n != len(blk) {
		return fmt.Errorf("didn't write the entire block")
	}

	return nil
}

func (bs *diskstore) Delete(key string) error {
	return bs.disk.Remove(key)
}
