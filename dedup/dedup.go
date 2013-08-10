package dedup

import (
	"crypto/sha1"
	"io"
)

const blkSize = 1024

type BlkInfo struct {
	Hash   []byte
	Length int
}

func GetBlocks(r io.Reader) (<-chan BlkInfo, <-chan error) {
	ret, errs := make(chan BlkInfo), make(chan error)

	go feed(r, ret, errs)

	return ret, errs
}

func feed(r io.Reader, blocks chan BlkInfo, errs chan error) {
	h := sha1.New()
	for {
		h.Reset()
		n, err := io.CopyN(h, r, blkSize)
		if err == io.EOF || err == nil {
			blocks <- BlkInfo{h.Sum(nil), int(n)}

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
