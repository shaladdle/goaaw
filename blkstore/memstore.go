package blkstore

import (
	"fmt"
)

type memstore struct {
	data map[string][]byte
}

func NewMemStore() BlkStore {
	return &memstore{make(map[string][]byte)}
}

func (bs *memstore) Get(key string) ([]byte, error) {
	val, ok := bs.data[key]
	if !ok {
		return nil, fmt.Errorf("data[%v] does not exist", key)
	}

	return val, nil
}

func (bs *memstore) Put(key string, blk []byte) error {
	bs.data[key] = blk
	return nil
}

func (bs *memstore) Delete(key string) error {
	_, ok := bs.data[key]
	if !ok {
		return fmt.Errorf("data[%v] does not exist", key)
	}

	delete(bs.data, key)
	return nil
}

func (bs *memstore) Size() int64 {
	var sum int64
	for _, v := range bs.data {
		sum += int64(len(v))
	}
	return sum
}
