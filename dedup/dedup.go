package dedup

type BlockInfo struct {
    Pos int64
    Size int
    Hash []byte
}

type Deduper interface {
    ComputeBlockList(path string) ([]BlockInfo, error)
}
