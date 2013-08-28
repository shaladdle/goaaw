// Package std provides an object satisfying the aaw/fs FileSystem interface
// that uses the standard library's os package.
package std

import (
	"os"
	"path"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

func New(root string) FileSystem {
    return FileSystem(root)
}

type FileSystem string

func (fs FileSystem) path(p string) string {
	return path.Join(string(fs), p)
}

func (fs FileSystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	info, err := os.Stat(fs.path(name))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fuse.ToAttr(info), fuse.OK
}

func (fs FileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(os.Chtimes(fs.path(name), *Atime, *Mtime))
}

func (fs FileSystem) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(os.Truncate(fs.path(name), int64(size)))
}

func (fs FileSystem) Access(name string, mode uint32, context *fuse.Context) fuse.Status {
	info, err := os.Stat(fs.path(name))
	if err != nil {
		return fuse.ToStatus(err)
	}

	if uint32(info.Mode())&mode != 0 {
		return fuse.OK
	}

	return fuse.EACCES
}

func (fs FileSystem) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	if err := os.Mkdir(fs.path(name), os.FileMode(mode)); err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.OK
}

// TODO: Syscall
func (fs FileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(syscall.Mknod(fs.path(name), mode, int(dev)))
}

func (fs FileSystem) Rename(oldName, newName string, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(os.Rename(fs.path(oldName), fs.path(newName)))
}

// TODO: Should I use the syscall here?
func (fs FileSystem) Rmdir(name string, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(os.Remove(fs.path(name)))
}

func (fs FileSystem) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	f, err := os.OpenFile(fs.path(name), int(flags), 0777)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return nodefs.NewLoopbackFile(f), fuse.OK
}

func (fs FileSystem) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	f, err := os.OpenFile(fs.path(name), int(flags), os.FileMode(mode))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return nodefs.NewLoopbackFile(f), fuse.OK
}

func (fs FileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	f, err := os.Open(fs.path(name))
	if err != nil {
		code = fuse.ToStatus(err)
		return
	}
	defer f.Close()

	infos, err := f.Readdir(0)
	if err != nil {
		code = fuse.ToStatus(err)
		return
	}

	stream = make([]fuse.DirEntry, len(infos))
	for i, info := range infos {
		stream[i] = fuse.DirEntry{
			Mode: uint32(info.Mode()),
			Name: info.Name(),
		}
	}

	return
}
