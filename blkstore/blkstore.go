package blkstore

// BlkStore represents a simple key value store.
type BlkStore interface {
	Get(key string) ([]byte, error)
	Put(key string, blk []byte) error
	Delete(key string) error
}

// BlkCache represents a cache with automatic eviction.
type BlkCache interface {
	Get(key string) ([]byte, error)
	Put(key string, blk []byte) error
	Has(key string) bool
}
