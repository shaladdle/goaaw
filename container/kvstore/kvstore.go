package kvstore

import (
	"io"
	"reflect"
)

type Coder interface {
	NewEncoder(io.Writer) Encoder
	NewDecoder(io.Reader) Decoder
}

type Encoder interface {
	EncodeValue(reflect.Value) error
}

type Decoder interface {
	DecodeValue(reflect.Value) error
}

type KVStore interface {
	Put(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
	Del(key interface{}) error
}
