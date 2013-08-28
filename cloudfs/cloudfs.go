package cloudfs

import (
	"fmt"
	"path"
	"time"

	remotestore "github.com/shaladdle/goaaw/fs/remote"
	anet "github.com/shaladdle/goaaw/net"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// Internal directory names
const (
	metaName    = "meta"
	stagingName = "staging"
)

// Xattr used for indicating whether something is big or not and the value
// that is saved in the actual attr.
const (
	xAttrBigName       = "isBig"
	xAttrBigValue byte = 1
)

type cloudFileSystem struct {
	debug    bool
	root     string
	hostport string

	meta    pathfs.FileSystem
	staging pathfs.FileSystem
	remote  *remotestore.Client
}

func New(root, hostport string) (pathfs.FileSystem, error) {
	// TODO: Do I have to call OnMount? What initialization is needed for these?
	meta := pathfs.NewLoopbackFileSystem(path.Join(root, metaName))
	staging := pathfs.NewLoopbackFileSystem(path.Join(root, stagingName))

	// TODO: Should this be done on mount instead of when instantiating the
	// object?
	remote, err := remotestore.NewClient(anet.TCPDialer(hostport))
	if err != nil {
		return nil, err
	}

	return &cloudFileSystem{
		root:     root,
		hostport: hostport,
		meta:     meta,
		staging:  staging,
		remote:   remote,
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

// TODO: What should we do about permissions? For now just returning OK and can
// figure out if we want to actually store some kind of permissions later. I
// think this would require making user accounts or something, which seems
// hard/stupid.
func (fs *cloudFileSystem) Chmod(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fuse.OK
}

func (fs *cloudFileSystem) Chown(name string, uid uint32, gid uint32, context *fuse.Context) fuse.Status {
	return fuse.OK
}

func (fs *cloudFileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) fuse.Status {
	return fs.meta.Utimens(name, Atime, Mtime, context)
}

func (fs *cloudFileSystem) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) Access(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Access(name, mode, context)
}

func (fs *cloudFileSystem) Link(oldName string, newName string, context *fuse.Context) fuse.Status {
	return fuse.EINVAL
}

func (fs *cloudFileSystem) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Mkdir(name, mode, context)
}

func (fs *cloudFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	return fs.meta.Mknod(name, mode, dev, context)
}

func (fs *cloudFileSystem) Rename(oldName string, newName string, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) Rmdir(name string, context *fuse.Context) fuse.Status {
	return fs.meta.Rmdir(name, context)
}

func (fs *cloudFileSystem) Unlink(name string, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) GetXAttr(name string, attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	if attribute == xAttrBigName {
		return nil, fuse.ToStatus(fmt.Errorf("%s is a reserved attribute name", attribute))
	}

	// TODO: I think this might not be good... Probably need to mount the meta
	// filesystem and then make calls using stdlib funcs.
	return fs.meta.GetXAttr(name, attribute, context)
}

func (fs *cloudFileSystem) ListXAttr(name string, context *fuse.Context) (attributes []string, code fuse.Status) {
	panic("not implemented")
}

func (fs *cloudFileSystem) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
}

func (fs *cloudFileSystem) OnUnmount() {
}

func (fs *cloudFileSystem) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	panic("not implemented")
}

func (fs *cloudFileSystem) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	panic("not implemented")
}

func (fs *cloudFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	return fs.meta.OpenDir(name, context)
}

func (fs *cloudFileSystem) Symlink(value string, linkName string, context *fuse.Context) fuse.Status {
	panic("not implemented")
}

func (fs *cloudFileSystem) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	panic("not implemented")
}

func (fs *cloudFileSystem) StatFs(name string) *fuse.StatfsOut {
	panic("not implemented")
}
