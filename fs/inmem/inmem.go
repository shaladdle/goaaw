// Package inmem provides an in memory file system.
package inmem

import (
    "io"
    "bytes"
    "fmt"
    "path"
    "strings"
    "os"
    "time"
)

const (
    rootName = "root"
    dirSep = "/"
)

func split(s string) []string {
    return strings.Split(s, dirSep)
}

type readCloserWrapper struct {
    io.Reader
}

func (readCloserWrapper) Close() error {
    return nil
}

func errFileNotFound(fpath string) error {
    return fmt.Errorf("file '%v' not found", fpath)
}

type fileInfo struct {
    name string
    size int64
    isDir bool
    modTime time.Time
}

func (fileInfo) Mode() os.FileMode {
    panic("Not implemented")
}

func (info fileInfo) Name() string {
    return info.name
}

func (info fileInfo) Size() int64 {
    return info.size
}

func (info fileInfo) IsDir() bool {
    return info.isDir
}

func (info fileInfo) Sys() interface{} {
    return nil
}

func (info fileInfo) ModTime() time.Time {
    return info.modTime
}

type dirNode struct {
    fileInfo
    children []os.FileInfo
}

func (d *dirNode) String() string {
    return fmt.Sprintf("('%s' %s)", d.Name(), d.children)
}

type node interface {
    os.FileInfo
    //Parent() node
}

type InMemFileSystem struct {
	data map[string][]byte
    nodes map[string]node
    root *dirNode
}

func New() *InMemFileSystem {
    rootNode := &dirNode{
        fileInfo: fileInfo{rootName, 0, true, time.Now()},
    }

    return &InMemFileSystem{
        data: make(map[string][]byte),
        nodes: map[string]node{rootName: rootNode},
        root: rootNode,
    }
}

// normPath normalizes the path passed in by prepending a '/' character and
// then executing path.Join.
func (fs *InMemFileSystem) normPath(tokens ...string) string {
    return path.Join(append([]string{rootName}, tokens...)...)
}

func (fs *InMemFileSystem) Open(fpath string) (io.ReadCloser, error) {
    buf, ok := fs.data[fpath]
    if !ok {
        return nil, errFileNotFound(fpath)
    }

    return readCloserWrapper{bytes.NewReader(buf)}, nil
}

type file struct {
    *bytes.Buffer
    fs *InMemFileSystem
    fpath string
}

func (fs *InMemFileSystem) newFile(fpath string) *file {
    return &file{&bytes.Buffer{}, fs, fpath}
}

func (f *file) Close() error {
    b := f.Buffer.Bytes()
    info := fileInfo{
        name: path.Base(f.fpath),
        size: int64(len(b)),
        modTime: time.Now(),
        isDir: false,
    }

    dpath := f.fs.normPath(path.Dir(f.fpath))

    // To make things simple, just call mkdir so we know the directory is 
    // set up.
    err := f.fs.mkdir(f.fs.root, split(dpath))
    if err != nil {
        return err
    }

    dchildren := f.fs.nodes[dpath].(*dirNode).children
    dchildren = append(dchildren, info)

    f.fs.data[f.fpath] = b
    f.fs.nodes[f.fpath] = info

    return nil
}

func (fs *InMemFileSystem) mkdir(n *dirNode, dirs []string) error {
    // TODO: Is this necessary? Not sure if we should be 'normalizing' things 
    // or not.
    if n.Name() == rootName {
        dirs = dirs[1:]
    }

    for _, c := range n.children {
        if c.Name() == dirs[0] {
            if c.IsDir() {
                if len(dirs[1:]) == 0 {
                    return nil
                }

                return fs.mkdir(c.(*dirNode), dirs[1:])
            } else {
                return fmt.Errorf("'%s' already exists as a regular file")
            }
        }
    }

    c := &dirNode{
        fileInfo: fileInfo{
            name: dirs[0],
            size: 0,
            isDir: true,
        },
        children: []os.FileInfo{},
    }

    n.children = append(n.children, c)

    fs.nodes[fs.normPath(dirs...)] = c

    if len(dirs) > 1 {
        return fs.mkdir(c, dirs[1:])
    }

    return nil
}

func (fs *InMemFileSystem) Mkdir(dpath string) error {
    return fs.mkdir(fs.root, split(dpath))
}

func (fs *InMemFileSystem) Create(fpath string) (io.WriteCloser, error) {
    return fs.newFile(fpath), nil
}

func (fs *InMemFileSystem) Stat(fpath string) (os.FileInfo, error) {
    var traverse func(*dirNode, []string) (os.FileInfo, error)
    traverse = func(n *dirNode, tokens []string) (os.FileInfo, error) {
        for _, c := range n.children {
            if c.Name() == tokens[0] {
                if len(tokens) == 1 {
                    // If we are at the last token, just return the fileInfo we 
                    // found.
                    switch c := c.(type) {
                    case *dirNode:
                        return c.fileInfo, nil
                    case fileInfo:
                        return c, nil
                    default:
                        panic("child of a node was not fileInfo and als not *dirNode")
                    }
                } else if c.IsDir() {
                    // If this is not the last token, and this child is a 
                    // directory, go deeper.
                    return traverse(c.(*dirNode), tokens[1:])
                } else {
                    // If it's not a directory, and we haven't exhausted our 
                    // tokens, now we give up.
                    return nil, errFileNotFound(fpath)
                }
            }
        }

        return nil, errFileNotFound(fpath)
    }

    return traverse(fs.root, split(fpath))
}

func (fs *InMemFileSystem) Remove(fpath string) error {
    return errFileNotFound(fpath)
}

func (fs *InMemFileSystem) GetFiles(fpath string) ([]os.FileInfo, error) {
    return nil, errFileNotFound(fpath)
}
