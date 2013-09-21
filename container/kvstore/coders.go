package kvstore

import (
	"encoding/gob"
	"io"
)

type gobCoder struct{}

func (gobCoder) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

func (gobCoder) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}
