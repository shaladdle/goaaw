package rpc

import (
	"encoding/gob"
	"fmt"
	"io"
)

var defaultCoder Coder = gobCoder{}

func init() {
	gob.Register(rpcClass(0))
}

type Coder interface {
	Encode(w io.Writer, src interface{}) error
	Decode(r io.Reader, dst interface{}) error
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

func (gobCoder) Decode(r io.Reader, dst interface{}) error {
	return gob.NewDecoder(byteReader{r}).Decode(dst)
}
