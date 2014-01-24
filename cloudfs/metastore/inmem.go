package metastore

import (
    "fmt"
)

type inMem map[string]bool

func NewInMem() MetaStore {
	return inMem(make(map[string]bool))
}

func (m inMem) IsBig(key string) (bool, error) {
	ret, exists := m[key]
	if !exists {
		return false, fmt.Errorf("key %v does not exist", key)
	}

	return ret, nil
}

func (m inMem) SetBig(key string, value bool) error {
	m[key] = value
	return nil
}

func (m inMem) Remove(key string) error {
	delete(m, key)
	return nil
}

func (inMem) Close() error {
	return nil
}
