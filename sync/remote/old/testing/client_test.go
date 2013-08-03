package testing

import (
	"aaw/sync/remote/client"
    "io"
	"aaw/sync/remote/common"
	"aaw/sync/remote/server"
	"aaw/sync/util"
	"fmt"
    "bytes"
	"os"
	"path"
	"testing"
    "crypto/sha1"
)

var (
	testdir    = "/tmp/testdir"
	testClidir = path.Join(testdir, "client")
	testSrvdir = path.Join(testdir, "server")
)

func runCommands(t *testing.T, s common.Storage, cmds []interface{}) {
    for _, cmd := range cmds {
        var err error

        switch cmd := cmd.(type) {
        case common.GetMsg:
            err = s.Get(cmd.Id)
        case common.PutMsg:
            err = s.Put(cmd.Id)
        case common.DelMsg:
            err = s.Delete(cmd.Id)
        default:
            t.Fatal("unrecognized command")
        }

        if err != nil {
            t.Fatalf("command '%v' failed: %v", cmd, err)
        }
    }
}

type testFile struct {
    name, contents string
}

func compareFiles(fpath1, fpath2 string) (bool, error) {
    f1, err := os.Open(fpath1)
    if err != nil {
        return false, nil
    }
    defer f1.Close()

    f2, err := os.Open(fpath2)
    if err != nil {
        return false, nil
    }
    defer f2.Close()

    h1Err := make(chan error)
    h1Result := make(chan []byte)
    go func() {
        h1 := sha1.New()
        _, err := io.Copy(h1, f1)
        if err != nil {
            h1Err <- err
        } else {
            h1Result <- h1.Sum(nil)
        }
    } ()

    h2 := sha1.New()
    _, err = io.Copy(h2, f2)
    if err != nil {
        return false, nil
    }

    select {
    case err := <-h1Err:
        return false, err
    case res1 := <-h1Result:
        return bytes.Equal(res1, h2.Sum(nil)), nil
    }

    return false, fmt.Errorf("unreachable code")
}

func testStorage(t *testing.T, s common.Storage, lpath, rpath string, files []testFile) {
    for _, file := range files {
        err := s.Put(file.name)
        if err != nil {
            t.Error(err)
        }

        lfile := path.Join(lpath, file.name)
        rfile := path.Join(rpath, file.name)
        same, err := compareFiles(lfile, rfile)
        if err != nil {
            t.Error(err)
        }

        if !same {
            t.Errorf("files '%v' and '%v' did not match", lfile, rfile)
        }
    }
}

func TestClient(t *testing.T) {
	srvHostport := "localhost:8000"

	dirs := []string{testClidir, testSrvdir}
	for _, d := range dirs {
		err := os.RemoveAll(d)
		if err != nil {
			t.Fatal(err)
		}

		err = util.TryMkdir(d)
		if err != nil {
			t.Fatal(err)
		}
	}

	writeStringFile := func(fpath, contents string) error {
		f, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer f.Close()

		n, err := f.Write([]byte(contents))
		if err != nil {
			return err
		}
		if n != len(contents) {
			return fmt.Errorf("didn't write complete string")
		}

		return nil
	}

    cmds := []interface{}{}

    cliFiles := []testFile{
        {"test1", "qpwoeifjqoreighln;nv;qh[g"},
    }
    for _, testf := range cliFiles {
        err := writeStringFile(
            path.Join(testClidir, testf.name),
            testf.contents,
        )
        if err != nil {
            t.Fatal(err)
        }

        cmds = append(cmds, common.PutMsg{testf.name})
    }

    srvFiles := []testFile{
        {"test2", "qpwoeifjwogvbndfa;sdfqh[g"},
    }
    for _, testf := range srvFiles {
        err := writeStringFile(
            path.Join(testSrvdir, testf.name),
            testf.contents,
        )
        if err != nil {
            t.Fatal(err)
        }

        cmds = append(cmds, common.GetMsg{testf.name})
    }

    _, err := server.New(srvHostport, testSrvdir)
	if err != nil {
		t.Fatal(err)
	}

	cli, err := client.New(srvHostport, testClidir)
	if err != nil {
		t.Fatal(err)
	}

    runCommands(t, cli, cmds)
}
