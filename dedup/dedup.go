package dedup

import (
	"bytes"
	"crypto/sha1"
	"io"
)

const blkSize = 1024

type BlkInfo struct {
	Hash    []byte
	Content []byte
}

func GetBlocks(r io.Reader) (<-chan BlkInfo, <-chan error) {
	ret, errs := make(chan BlkInfo), make(chan error)
	go feed(r, ret, errs)
	return ret, errs
}

func feed(r io.Reader, blocks chan BlkInfo, errs chan error) {
	h := sha1.New()
	for {
		buf := &bytes.Buffer{}
		h.Reset()
		w := io.MultiWriter(h, buf)
		_, err := io.CopyN(w, r, blkSize)
		if err == io.EOF || err == nil {
			blocks <- BlkInfo{h.Sum(nil), buf.Bytes()}

			if err == io.EOF {
				close(blocks)
				close(errs)
				break
			}
		} else if err != nil {
			close(blocks)
			errs <- err
			close(errs)
			break
		}
	}
}
