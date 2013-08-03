package remote

import (
	"log"
	"path"
	"testing"

	"aaw/sync/util"
)

func TestPutGet(t *testing.T) {
	hostport := "127.0.0.1:9090"

	srvDir := path.Join(util.GetTestPath(), "server")
	util.TryMkdir(srvDir)

	_, err := NewServer(hostport, srvDir)
	if err != nil {
		t.Fatal(err)
	}

	cli, err := NewClient(hostport)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := util.GenRandFiles([]util.FileSpec{
		{5, 10 * util.MB},
	})
	if err != nil {
		t.Fatal(err)
	}

	log.Println(paths)
	for _, fname := range paths {
		err := cli.Put(fname, path.Join(util.GetTestPath(), fname))
		if err != nil {
			t.Fatal(err)
		}
	}
}
