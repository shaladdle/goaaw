package main

import (
	"flag"
	"log"

	"github.com/shaladdle/goaaw/cloudfs"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

var (
	root  = flag.String("root", "./cloudfs-internal", "Root directory for cloudfs data structures")
	raddr = flag.String("remote", "localhost:9000", "IP address of remote storage")
	mnt   = flag.String("mnt", "./cloudfs", "Mount point for the filesystem")
)

func main() {
	flag.Parse()

	fs, err := cloudfs.New(*root, *raddr)
	if err != nil {
		log.Fatal("Cloudfs initialization error: ", err)
	}

	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountFileSystem(*mnt, nfs, nil)
	if err != nil {
		log.Fatal("Mount fail: ", err)
	}
	server.Serve()
}
