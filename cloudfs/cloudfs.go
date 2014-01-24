package cloudfs

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/shaladdle/goaaw/cloudfs/metastore"
	remotestore "github.com/shaladdle/goaaw/filestore/remote"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// Internal directory names
const (
	metaName    = "meta"
	stagingName = "staging"
)

type cloudFileSystem struct {
	debug    bool
	root     string
	hostport string

	metadb  metastore.MetaStore
	meta    pathfs.FileSystem
	staging pathfs.FileSystem
	remote  blkstore.BlkStore
}

func New(root, hostport string) (pathfs.FileSystem, error) {
	return &cloudFileSystem{
		root:     root,
		hostport: hostport,
		meta:     pathfs.NewLoopbackFileSystem(path.Join(root, metaName)),
		staging:  pathfs.NewLoopbackFileSystem(path.Join(root, stagingName)),
        remote:   blkstore.NewMemStore()
	}, nil
}

func (fs *cloudFileSystem) String() string {
	return fmt.Sprintf("cloudfs{root: %v, hostport: %v}", fs.root, fs.hostport)
}

func (fs *cloudFileSystem) SetDebug(debug bool) {
	fs.debug = debug
}

func (fs *cloudFileSystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return fs.meta.GetAttr(name, context)
}

func (fs *cloudFileSystem) Chmod(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Chmod(name, mode, context)
}

func (fs *cloudFileSystem) Chown(name string, uid uint32, gid uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Chown(name, uid, gid, context)
}

func (fs *cloudFileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) fuse.Status {
	return fs.meta.Utimens(name, Atime, Mtime, context)
}

func (fs *cloudFileSystem) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	return fs.meta.Truncate(name, size, context)
}

func (fs *cloudFileSystem) Access(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Access(name, mode, context)
}

func (fs *cloudFileSystem) Link(oldName string, newName string, context *fuse.Context) fuse.Status {
	return fs.meta.Link(oldName, newName, context)
}

func (fs *cloudFileSystem) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Mkdir(name, mode, context)
}

func (fs *cloudFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Mknod(name, mode, dev, context)
}

func (fs *cloudFileSystem) Rename(oldName string, newName string, context *fuse.Context) fuse.Status {
	return fs.meta.Rename(oldName, newName, context)
}

func (fs *cloudFileSystem) Rmdir(name string, context *fuse.Context) fuse.Status {
	return fs.meta.Rmdir(name, context)
}

func (fs *cloudFileSystem) Unlink(name string, context *fuse.Context) fuse.Status {
	return fuse.ENOSYS
}

func (fs *cloudFileSystem) GetXAttr(name string, attr string, context *fuse.Context) (data []byte, code fuse.Status) {
	return fs.meta.GetXAttr(name, attr, context)
}

func (fs *cloudFileSystem) ListXAttr(name string, context *fuse.Context) (attributes []string, code fuse.Status) {
	return fs.meta.ListXAttr(name, context)
}

func (fs *cloudFileSystem) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	return fs.meta.RemoveXAttr(name, attr, context)
}

func (fs *cloudFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	return fs.meta.SetXAttr(name, attr, data, flags, context)
}

func (fs *cloudFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
	// TODO: Add remote initialization when that's ready
	/*
		remote, err := remotestore.NewClient(anet.TCPDialer(hostport))
		if err != nil {
			return nil, err
		}
	*/

	if err := initLocal(fs.root); err != nil {
		log.Fatal("Error initializing private dir: ", err)
	}

	metadb, err := metastore.NewMetaDB(fs.root)
	if err != nil {
		log.Fatal("Error instantiating MetaDB: ", err)
	}

	fs.metadb = metadb
	//fs.remote = remote
}

func (fs *cloudFileSystem) OnUnmount() {
}

func (fs *cloudFileSystem) fileExists(name string) bool {
    return false
}

func (fs *cloudFileSystem) fileIsBig(name string) bool {
    return false
}

func (fs *cloudFileSystem) readFileMetadata(name string) (interface{}, error) {
    return nil, fmt.Errorf("NOT IMPLEMENTED")
}

func (fs *cloudFileSystem) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
    return fs.meta.Open(name, flags, context)
    /*
    if !fs.fileExists(name) {
        // The file doesn't exist yet, so the right thing to do is just return
        // f and status as they are.
        // TODO: Should these kinds of files go in the map?
        return fs.meta.Open(name, flags, context)
    } else if _, ok := openFiles[name]; ok {
        // If the file is open (check using some kind of hash table) we will
        // update some kind of reference count.
    } else if fs.fileIsBig(name) {
        // The file is not open and it's a big file, so we want to reconstruct the file by
        // downloading its blocks to a file in the scratch directory.
        info, err := fs.readFileMetadata(name)
        if err != nil {
            return 
        }
    } else {
        // The file is small and it's not already open, so we just open it.
        return fs.meta.Open(name, flags, context)
    }
    */
}

func (fs *cloudFileSystem) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	return fs.meta.Create(name, flags, mode, context)
}

func (fs *cloudFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	return fs.meta.OpenDir(name, context)
}

func (fs *cloudFileSystem) Symlink(value string, linkName string, context *fuse.Context) fuse.Status {
	return fs.meta.Symlink(value, linkName, context)
}

func (fs *cloudFileSystem) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	return fs.meta.Readlink(name, context)
}

func (fs *cloudFileSystem) StatFs(name string) *fuse.StatfsOut {
	return fs.meta.StatFs(name)
}
