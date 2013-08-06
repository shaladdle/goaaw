package blkstore

// BlkStore represents a simple key value store.
type BlkStore interface {
	Get(key string) ([]byte, error)
	Put(key string, blk []byte) error
	Delete(key string) error
}
