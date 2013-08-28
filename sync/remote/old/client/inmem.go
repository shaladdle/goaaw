package client

import (
	"github.com/shaladdle/goaaw/sync/remote/common"
	"bytes"
	"fmt"
	"io"
)

func NewInMemory() (*InMemStorage, error) {
	return &InMemStorage{make(map[string]*bytes.Buffer)}, nil
}

type InMemStorage struct {
	data map[string]*bytes.Buffer
}

func (ls *InMemStorage) Get(id string) (io.Reader, error) {
	if elem, exists := ls.data[id]; exists {
		tmp := *elem
		return &tmp, nil
	}

	return nil, fmt.Errorf("file '%v' does not exist", id)
}

func (ls *InMemStorage) Put(id string, r io.Reader) error {
	b := &bytes.Buffer{}

	_, err := io.Copy(b, r)
	if err != nil {
		return err
	}

	ls.data[id] = b

	return nil
}

func (ls *InMemStorage) Delete(id string) error {
	if _, exists := ls.data[id]; exists {
		delete(ls.data, id)
		return nil
	}

	return fmt.Errorf("file '%v' does not exist.", id)
}

func (ls *InMemStorage) GetIndex() (map[string]common.FileInfo, error) {
	ret := make(map[string]common.FileInfo)
	for k, v := range ls.data {
		ret[k] = common.FileInfo{
			Id:   k,
			Size: int64(v.Len()),
		}
	}

	return ret, nil
}
