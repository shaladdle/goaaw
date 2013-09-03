// Package std provides an object satisfying the aaw/fs FileSystem interface
// that uses the standard library's os package.
package std

import (
	"fmt"
	"io"
	"os"
	"path"
)

// FileSystem uses the os file operations to emulate a file system mounted
// at root.
type FileSystem struct {
	root string
}

func New(root string) FileSystem {
	return FileSystem{root}
}

func (fs FileSystem) Open(fpath string) (io.ReadCloser, error) {
	return os.Open(path.Join(fs.root, fpath))
}
func (fs FileSystem) Create(fpath string) (io.WriteCloser, error) {
	return os.Create(path.Join(fs.root, fpath))
}
func (fs FileSystem) Mkdir(dpath string) error { return os.MkdirAll(path.Join(fs.root, dpath), 0777) }
func (fs FileSystem) Stat(fpath string) (os.FileInfo, error) {
	return os.Stat(path.Join(fs.root, fpath))
}
func (fs FileSystem) Remove(fpath string) error { return os.Remove(path.Join(fs.root, fpath)) }

func (fs FileSystem) GetFiles(fpath string) ([]os.FileInfo, error) {
	fspath := path.Join(fs.root, fpath)
	info, err := os.Stat(fspath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%v is not a directory", fspath)
	}

	d, err := os.Open(fspath)
	if err != nil {
		return nil, err
	}

	fi, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}

	ret := make([]os.FileInfo, len(fi))

	for i, fi := range fi {
		if fi.Mode().IsRegular() {
			ret[i] = fi
		}
	}

	return ret, nil
}
