package fdcache

import (
	"github.com/shaladdle/goaaw/lrucache"
	"os"
)

type FdCache struct {
	lruCache *lrucache.LruCache
}

func NewFdCache(maxFds int) *FdCache {
	onEvict := func(value interface{}) {
		f := value.(*File)
		if f.closed {
			return
		}

		if err := f.Close(); err != nil {
			panic(err)
		}
	}

	return &FdCache{
		lrucache.New(maxFds, onEvict),
	}
}

func (c *FdCache) Open(fpath string) (*File, error) {
	return &File{
		cache:  c,
		path:   fpath,
		file:   nil,
		pos:    0,
		closed: true,
	}, nil
}

func (c *FdCache) requestOpen(f *File) error {
	// Do book-keeping by talking to the cache
	if !c.lruCache.Contains(f.path) {
		c.lruCache.Put(f.path, f)
	} else {
		c.lruCache.Get(f.path)
	}

	if f.closed {
		if err := f.Open(); err != nil {
			return err
		}
	}

	return nil
}

type File struct {
	cache  *FdCache
	path   string
	file   *os.File
	pos    int64
	closed bool
}

func (f *File) Open() error {
	var err error
	f.file, err = os.Open(f.path)
	if err != nil {
		return err
	}

	_, err = f.file.Seek(f.pos, 0)
	if err != nil {
		return err
	}

	f.closed = false

	return nil
}

func (f *File) Close() error {
	pos, err := f.file.Seek(0, 1)
	if err != nil {
		return err
	}

	f.pos = pos

	err = f.file.Close()
	if err != nil {
		return err
	}

	f.file = nil
	f.closed = true

	return nil
}

func (f *File) Read(p []byte) (int, error) {
	if f.closed {
		// Tell the cache we are going to open this file again, it will close
		// another file to make room
		err := f.cache.requestOpen(f)
		if err != nil {
			return 0, err
		}
	}

	return f.file.Read(p)
}
