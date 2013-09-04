package main

import (
	"flag"
	"log"

	"github.com/shaladdle/goaaw/cloudfs"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

var (
	data  = flag.String("data", "./cloudfs-internal", "Data directory for cloudfs internal structures")
	raddr = flag.String("remote", "localhost:9000", "IP address of remote storage")
	mnt   = flag.String("mnt", "./cloudfs", "Mount point for the filesystem")
)

func main() {
	flag.Parse()

	fs, err := cloudfs.New(*data, *raddr)
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
