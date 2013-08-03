package common

type Storage interface {
	Get(id string) error
	Put(id string) error
	Delete(id string) error
	GetIndex() (map[string]FileInfo, error)
}
