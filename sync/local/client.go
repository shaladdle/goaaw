package local

import (
	"aaw/sync/remote"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Index map[string]os.FileInfo

func createIndex(path string) (Index, error) {
	ret := make(Index)

	err := filepath.Walk(path, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ret[fpath] = info

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func watch(path string, done <-chan int, status <-chan Index) {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-done:
			ticker.Stop()
			break
		case <-ticker.C:
            createIndex(path)
		}
	}
}

type getMessage struct {
	id  string
	err chan error
}

type Client struct {
	remStorage  remote.Storage
	fatal       chan error
	status      <-chan Index
	killWatcher chan int
	get         chan getMessage
	getDone     chan string
	cache       *fileCache
	lpath       string
}

func (cli *Client) Fatal() <-chan error {
	return cli.fatal
}

func (cli *Client) Notify(id string, args ...string) error {
	if len(args) > 0 {
		msg := getMessage{
			args[0],
			make(chan error),
		}
		cli.get <- msg
		return <-msg.err
	}

	return fmt.Errorf("Not enough arguments for command", id)
}

func New(lpath string, cacheSize int64, remStorage remote.Storage) (*Client, error) {
	ret := &Client{
		status:      make(chan Index),
		killWatcher: make(chan int),
		get:         make(chan getMessage),
		getDone:     make(chan string),
		lpath:       lpath,
		cache:       newCache(cacheSize),
		remStorage:  remStorage,
	}

	go watch(lpath, ret.killWatcher, ret.status)
	go ret.director()

	return ret, nil
}

func (cli *Client) getFile(id string) {
	defer func() {
		cli.getDone <- id
	}()

	r, err := cli.remStorage.Get(id)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(cli.fullPath(id))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		panic(err)
	}
}

func (cli *Client) fullPath(id string) string {
	return path.Join(cli.lpath, id)
}

func (cli *Client) cacheGet(id string) {
	fullPath := cli.fullPath(id)
	info, err := os.Stat(fullPath)
	if err != nil {
		panic(err)
	}

	cli.cache.Get(id, info.Size())
}

func (cli *Client) director() {
	waitingGets := make(map[string]bool)
	for {
		select {
		case msg := <-cli.get:
			fmt.Println("Getting")
			if _, exists := waitingGets[msg.id]; exists {
				msg.err <- fmt.Errorf("Already getting %v", msg.id)
			} else {
				if !cli.cache.Contains(msg.id) {
					go cli.getFile(msg.id)
				} else {
					cli.cacheGet(msg.id)
				}

				msg.err <- nil
			}
		case id := <-cli.getDone:
			cli.cacheGet(id)
			delete(waitingGets, id)
		}
	}
}
