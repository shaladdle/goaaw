package dedupfs

import (
	"io"
	"os"
	"path"
	"superbox/util"
)

const (
	BLOCK_SIZE    = 4096
	INDEX_NAME    = "index"
	BLOCKS_FOLDER = "blocks"
)

type block struct {
	Hash string
	Length int
	Count int
}

type fileIndex struct {
	Files  map[string][]uint64 // maps paths to block ids
	Blocks blockMap // maps block id to block object
}

func (fi *fileIndex) getFileLength(fpath string) int64 {
	blockids := fi.Files[fpath]
	nfull := len(blockids) - 1
	lastIdx := nfull

	lastblock := fi.Blocks[blockids[lastIdx]]

	return int64(BLOCK_SIZE * int64(nfull)) + int64(lastblock.Length)
}

type FileSystem struct {
	rootPath string
	index    *fileIndex
}

func NewFileSystem(rootPath string) (*FileSystem, error) {
	var err error
	fs := &FileSystem{
		rootPath: rootPath,
		index: &fileIndex{},
	}

	err = fs.setupDirs()
	if err != nil {
		return nil, err
	}

	fs.index = &fileIndex{
		Files: make(map[string][]uint64),
		Blocks: make(blockMap),
	}
	indexPath := fs.getIndexPath()
	if fileExists(indexPath) {
		err = util.ReadConfig(indexPath, fs.index)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (fs *FileSystem) setupDirs() error {
	if !fileExists(fs.rootPath) {
		err := os.MkdirAll(fs.rootPath, 0766)
		if err != nil {
			return err
		}
	}

	blocksFolderPath := path.Join(fs.rootPath, BLOCKS_FOLDER)
	if !fileExists(blocksFolderPath) {
		err := os.MkdirAll(blocksFolderPath, 0766)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileSystem) Create(fpath string) (io.ReadWriteCloser, error) {
	if _, ok := fs.index.Files[fpath]; ok {
		err := fs.Remove(fpath)
		if err != nil {
			return nil, err
		}
	}

	blocks := []uint64{}
	fs.index.Files[fpath] = blocks

	ret := &FileHandle{
		length: 0,
		blocks: blocks,
		path: fpath,
	}

	return ret, nil
}

func (fs *FileSystem) Open(fpath string) (io.ReadWriteCloser, error) {
	if _, ok := fs.index.Files[fpath]; !ok {
		return fs.Create(fpath)
	}

	ret := &FileHandle{
		length: fs.index.getFileLength(fpath),
		blocks: fs.index.Files[fpath],
		path: fpath,
	}

	return ret, nil
}

// Basically all we have to do here is lower the count of each block 
// and remove that block from disk if this was the last reference to it
func (fs *FileSystem) Remove(fpath string) error {
	blocks := fs.index.Files[fpath]
	for bid := range blocks {
		ubid := uint64(bid)
		blk := fs.index.Blocks[ubid]
		if blk.Count--; blk.Count == 0 {
			err := fs.deleteBlock(ubid)
			if err != nil {
				return err
			}
		}
	}

	delete(fs.index.Files, fpath)

	return nil
}

func (fs *FileSystem) Close() error {
	return util.WriteConfig(fs.getIndexPath(), fs.index)
}

func (fs *FileSystem) deleteBlock(bid uint64) error {
	err := os.Remove(fs.getBlockPath(bid))
	if err != nil {
		return err
	}

	delete(fs.index.Blocks, bid)

	return nil
}

func (fs *FileSystem) getBlockPath(bid uint64) string {
	return path.Join(fs.rootPath, fs.index.Blocks[bid].Hash)
}

func (fs *FileSystem) getIndexPath() string {
	return path.Join(fs.rootPath, INDEX_NAME)
}

type FileHandle struct {
	length int64
	blocks []uint64
	path string
	position int64
	curBlock []byte
}

func (fh *FileHandle) Read(b []byte) (int, error) {
	numblocks := (len(b) / BLOCK_SIZE)
	if ((len(b) % BLOCK_SIZE) != 0 {
	}
	nleft := len(b)
	for i := 0; i < numblocks; i++ {
		os.Open(
	}
	return 0, nil
}

func (fh *FileHandle) Write(b []byte) (int, error) {
	return 0, nil
}

func (fh *FileHandle) Close() error {
	return nil
}
