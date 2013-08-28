// Package fs defines the basic filesystem interface.
package fs

import (
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

// TODO: Extended attributes would be nice here.

// FileSystem defines the basic interface of a filesystem. In the future this
// should be usable as with FUSE, so things like permissions will need to be
// added.
type FileSystem interface {
	GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status)
	Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) fuse.Status
	Truncate(name string, size uint64, context *fuse.Context) fuse.Status
	Access(name string, mode uint32, context *fuse.Context) fuse.Status
	Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status
	Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status
	Rename(oldName, newName string, context *fuse.Context) fuse.Status
	Rmdir(name string, context *fuse.Context) fuse.Status
	Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status)
	Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status)
	OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status)
}
