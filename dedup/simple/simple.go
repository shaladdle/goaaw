package simple

import (
    "io"
    "os"
    "crypto/sha1"

    "github.com/shaladdle/goaaw/dedup"
)

type deduper struct {
    blockSize int
}

func NewDeduper(blockSize int) dedup.Deduper {
    return &deduper{blockSize}
}

func (d *deduper) ComputeBlockList(path string) ([]dedup.BlockInfo, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }

    numBlocks := int(info.Size() / int64(d.blockSize))
    rem := int(info.Size() % int64(d.blockSize))

    if rem != 0 {
        numBlocks++
    }

    h := sha1.New()

    ret := make([]dedup.BlockInfo, numBlocks)
    for i := range ret {
        n, err := io.CopyN(h, f, int64(d.blockSize))
        if err != nil && err != io.EOF {
            return nil, err
        }

        ret[i].Pos = int64(i) * int64(d.blockSize)
        ret[i].Size = int(n)
        ret[i].Hash = h.Sum(nil)

        h.Reset()
    }

    return ret, nil
}
