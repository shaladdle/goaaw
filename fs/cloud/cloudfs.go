package cloud

import (
	"fmt"
	"io"
	"os"
	"path"

	"aaw/blkstore"
	"aaw/dedup"
	"aaw/fs"
	"aaw/fs/std"
	anet "aaw/net"
)

func hashStr(hash []byte) string { return fmt.Sprintf("%02x", hash) }

type vector struct {
	s []string
	m map[string]interface{}
}

func newVector() *vector {
	return &vector{
		m: make(map[string]interface{}),
	}
}

func (v *vector) AppendKey(key string) {
	v.s = append(v.s, key)
	v.m[key] = true
}

func (v *vector) Append(key string, val interface{}) {
	v.s = append(v.s, key)
	v.m[key] = val
}

func (v vector) Get(key string) interface{} {
	return v.m[key]
}

func (v vector) Has(key string) bool {
	_, has := v.m[key]
	return has
}

func (v vector) KeySlice() []string {
	return v.s
}

func (v vector) KeyIter() <-chan string {
	ch := make(chan string)
	go func() {
		for _, s := range v.s {
			ch <- s
		}
	}()
	return ch
}

func (v vector) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for _, key := range v.s {
			ch <- v.m[key]
		}
	}()
	return ch
}

func (a *vector) Diff(b *vector) []string {
	ret := []string{}
	for el := range a.Iter() {
		str := el.(string)
		if !b.Has(str) {
			ret = append(ret, str)
		}
	}
	return ret
}

func (a *vector) BiDiff(b *vector) ([]string, []string) {
	diffOneWay := func(v1, v2 *vector, list chan []string) {
		list <- v1.Diff(v2)
	}
	anotb, bnota := make(chan []string), make(chan []string)
	go diffOneWay(a, b, anotb)
	go diffOneWay(b, a, bnota)
	return <-anotb, <-bnota
}

const (
	cacheDirName   = "cache"
	stagingDirName = "staging"
)

type blkInfo struct {
	hash []byte
	refs int
}

type FileSystem struct {
	// Staging area for files being read/written.
	staging fs.FileSystem

	// Cache object used to avoid calls to the cloud.
	cache blkstore.BlkCache

	// Interface to cloud storage
	cloud blkstore.BlkStore

	// Information about a file, keyed by block hash.
	blkInfo map[string]blkInfo

	// List of blocks that make up a file, keyed by relative path.
	blkLists map[string]*vector
}

type File struct {
	io.WriteCloser
	fpath string
	fs    *FileSystem
}

func (f *File) Close() error {
	if err := f.WriteCloser.Close(); err != nil {
		return err
	}

	if err := f.fs.commit(f.fpath); err != nil {
		return err
	}

	return nil
}

// NewDefault creates a new cloudfs client. Currently TCP is used, so a string
// containing hostname and port number (like 'localhost:9000') is required.
// Before calling this function, an aaw/fs/remote server must be instantiated
// at the server machine.
func NewDefault(hostport, root string, cacheSize int64) (*FileSystem, error) {
	cloud, err := blkstore.NewRemoteStore(anet.TCPDialer(hostport))
	if err != nil {
		return nil, err
	}

	cacheDisk := blkstore.NewDiskStore(path.Join(root, cacheDirName))
	return NewFileSystem(
		std.New(path.Join(root, stagingDirName)),
		cloud,
		cacheDisk,
		cacheSize,
	), nil
}

func NewFileSystem(staging fs.FileSystem, cloud, cacheStore blkstore.BlkStore, cacheSize int64) *FileSystem {
	return &FileSystem{
		staging:  staging,
		cache:    blkstore.NewCache(cacheStore, cacheSize),
		cloud:    cloud,
		blkInfo:  make(map[string]blkInfo),
		blkLists: make(map[string]*vector),
	}
}

// retrieveBlock gets the block from the remote (or from the cache if the
// cache has it) and writes it to w.
func (fs *FileSystem) retrieveBlock(hash []byte, w io.Writer) error {
	var (
		blk []byte
		err error
	)

	hstr := hashStr(hash)
	if !fs.cache.Has(hstr) {
		if blk, err = fs.cloud.Get(hstr); err != nil {
			return err
		}

		if err = fs.cache.Put(hstr, blk); err != nil {
			return err
		}
	} else {
		if blk, err = fs.cache.Get(hstr); err != nil {
			return err
		}
	}

	if n, err := w.Write(blk); err != nil {
		return err
	} else if n != len(blk) {
		return fmt.Errorf("partial block written")
	}

	return nil
}

func (fs *FileSystem) commit(fpath string) error {
	f, err := fs.staging.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	oldBlocks, ok := fs.blkLists[fpath]
	newBlocks := newVector()
	blks, errs := dedup.GetBlocks(f)
	for info := range blks {
		hstr := hashStr(info.Hash)
		if ok && oldBlocks.Has(hstr) {
			continue
		}
		if err := fs.incBlk(info); err != nil {
			return err
		}
		newBlocks.AppendKey(hstr)
	}
	for err := range errs {
		return err
	}

	fs.blkLists[fpath] = newBlocks

	if ok {
		dec := oldBlocks.Diff(newBlocks)
		for _, hstr := range dec {
			if err := fs.decBlk(hstr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fs *FileSystem) reconstruct(f io.Writer, blkList []string) error {
	for _, hstr := range blkList {
		var (
			blk []byte
			err error
		)

		if fs.cache.Has(hstr) {
			blk, err = fs.cache.Get(hstr)
			if err != nil {
				return err
			}
		} else {
			blk, err = fs.cloud.Get(hstr)
			if err != nil {
				return err
			}

			err = fs.cache.Put(hstr, blk)
			if err != nil {
				return err
			}
		}

		if _, err = f.Write(blk); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileSystem) incBlk(info dedup.BlkInfo) error {
	hstr := hashStr(info.Hash)
	if curInfo, ok := fs.blkInfo[hstr]; !ok {
		fs.blkInfo[hstr] = blkInfo{info.Content, 1}

		if err := fs.cache.Put(hstr, info.Content); err != nil {
			return err
		}

		if err := fs.cloud.Put(hstr, info.Content); err != nil {
			return err
		}
	} else {
		curInfo.refs++
		fs.blkInfo[hstr] = curInfo
	}

	return nil
}

func (fs *FileSystem) decBlk(hstr string) error {
	if curInfo, ok := fs.blkInfo[hstr]; !ok {
		return fmt.Errorf("file does note exist")
	} else {
		curInfo.refs--
		fs.blkInfo[hstr] = curInfo
	}

	return nil
}

func (fs *FileSystem) Open(fpath string) (io.ReadCloser, error) {
	if _, ok := fs.blkLists[fpath]; !ok {
		return nil, fmt.Errorf("file not found")
	}

	f, err := fs.staging.Create(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := fs.reconstruct(f, fs.blkLists[fpath].KeySlice()); err != nil {
		return nil, err
	}
	// reconstruct the file in the staging directory and return a handle to it
	return fs.staging.Open(fpath)
}

func (fs *FileSystem) Create(fpath string) (io.WriteCloser, error) {
	f, err := fs.staging.Create(fpath)
	if err != nil {
		return nil, err
	}

	return &File{f, fpath, fs}, nil
}

func (fs *FileSystem) Mkdir(fpath string) error {
	return fmt.Errorf("not implemented")
}

func (fs *FileSystem) Stat(fpath string) (os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fs *FileSystem) Remove(fpath string) error {
	return fmt.Errorf("not implemented")
}

func (fs *FileSystem) GetFiles(fpath string) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
