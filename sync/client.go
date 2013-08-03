package main

import (
	"aaw/sync/local"
	"aaw/sync/remote"
	"aaw/sync/ui"
	"aaw/sync/util"
	"bytes"
	"flag"
	"fmt"
	"os"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

var (
	localdir = flag.String("ldir", "/tmp/localdir", "")
)

func main() {
	flag.Parse()

	if err := util.TryMkdir(*localdir); err != nil {
		fmt.Println("Error ensuring directory was created:", err)
		return
	}

	storage, err := remote.NewInMemoryStorage()
	if err != nil {
		fmt.Println("Error starting remote storage:", err)
		return
	}

	b := bytes.NewBufferString("hi there")
	if err := storage.Put("test", b); err != nil {
		fmt.Println("Error putting test file:", err)
		return
	}

	cli, err := local.New(*localdir, 10*GB, storage)
	if err != nil {
		fmt.Println("Error starting client:", err)
		return
	}

	events := []ui.EventInfo{
		{"get", 1},
		{"put", 1},
		{"delete", 1},
	}
	ui, err := ui.New(cli, events)
	if err != nil {
		fmt.Println("Error starting ui:", err)
		return
	}

	select {
	case err := <-ui.Fatal():
		fmt.Println("Unexpected UI shutdown:", err)
	case err := <-cli.Fatal():
		fmt.Println("Unexpected client shutdown:", err)
	}
}
