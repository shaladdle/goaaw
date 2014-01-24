package metastore

type MetaStore interface {
	IsBig(key string) (bool, error)
	SetBig(key string, value bool) error
	Remove(key string) error
	Close() error
}
