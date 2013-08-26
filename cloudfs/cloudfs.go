package cloudfs

import (
    "aaw/fs/remote"

    "github.com/hanwen/go-fuse/fuse"
)

type cloudFileSystem struct {
    meta fuse.FileSystem
    staging fuse.FileSystem
    remote *remotestore.Client
}

func New() (fuse.FileSystem, error) {
    return cloudFileSystem{
    }
}

type 
