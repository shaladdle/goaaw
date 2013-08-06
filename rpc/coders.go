package rpc

import (
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
)

var defaultCoder Coder = gobCoder{}

func init() {
	gob.Register(rpcClass(0))
	gob.Register(ErrNil)
}

type Coder interface {
	Encode(w io.Writer, src interface{}) error
	EncodeValue(w io.Writer, src reflect.Value) error
	Decode(r io.Reader, dst interface{}) error
	DecodeValue(r io.Reader, dst reflect.Value) error
}

type gobCoder struct{}

type byteReader struct {
	io.Reader
}

func (br byteReader) ReadByte() (byte, error) {
	b := make([]byte, 1)

	if n, err := br.Read(b); err != nil {
		return 0, err
	} else if n == 0 {
		return 0, fmt.Errorf("no bytes read")
	}

	return b[0], nil
}

func (gobCoder) Encode(w io.Writer, src interface{}) error {
	return gob.NewEncoder(w).Encode(src)
}

func (gobCoder) EncodeValue(w io.Writer, src reflect.Value) error {
	return gob.NewEncoder(w).EncodeValue(src)
}

func (gobCoder) Decode(r io.Reader, dst interface{}) error {
	if _, ok := r.(io.ByteReader); !ok {
		return gob.NewDecoder(byteReader{r}).Decode(dst)
	}

	return gob.NewDecoder(r).Decode(dst)
}

func (gobCoder) DecodeValue(r io.Reader, dst reflect.Value) error {
	if _, ok := r.(io.ByteReader); !ok {
		return gob.NewDecoder(byteReader{r}).DecodeValue(dst)
	}

	return gob.NewDecoder(r).DecodeValue(dst)
}
