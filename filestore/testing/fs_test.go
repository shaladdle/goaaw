package testing

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/shaladdle/goaaw/filestore"
	"github.com/shaladdle/goaaw/filestore/remote"
	"github.com/shaladdle/goaaw/filestore/std"
	"github.com/shaladdle/goaaw/filestore/util"
	"github.com/shaladdle/goaaw/testutil"
)

type testInfo struct {
	name  string
	setup func(*testing.T) (fs.FileStore, func(), error)
}

var tests = []testInfo{
	{"std", func(t *testing.T) (fs.FileStore, func(), error) {
		te := testutil.NewTestEnv("testcase-stdfs", t)
		return std.New(te.Root()), func() { te.Teardown() }, nil
	}},
	{"remote", func(t *testing.T) (fs.FileStore, func(), error) {
		const hostport = "localhost:9000"

		te := testutil.NewTestEnv("testcase-net-remotefs", t)

		srv, err := remote.NewTCPServer(te.Root(), hostport)
		if err != nil {
			return nil, nil, err
		}

		cli, err := remote.NewTCPClient(hostport)
		if err != nil {
			return nil, nil, err
		}

		cleanup := func() {
			srv.Close()
			cli.Close()
			te.Teardown()
		}

		return cli, cleanup, nil
	}},
}

func checkFileInfo(t *testing.T, testName string, got, want os.FileInfo) {
	if got, want := got.Name(), want.Name(); got != want {
		t.Errorf("test %v: name mismatch, got %v, want %v", testName, got, want)
	}
	if got, want := got.Size(), want.Size(); got != want {
		t.Errorf("test %v: size is incorrect, got %v, want 0", testName, got)
	}
	if got, want := got.IsDir(), want.IsDir(); got != want {
		t.Errorf("test %v: IsDir is incorrect, got %v, want false", testName, got)
	}
}

// TestWriteRead writes a file and then reads it back and compares the hash to
// make sure the same contents was read and written.
func TestWriteRead(t *testing.T) {
	const readTimes = 5

	te := testutil.NewTestEnv("TestWriteRead", t)
	defer te.Teardown()

	fname := "test"

	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		w, err := fs.Create(fname)
		if err != nil {
			t.Errorf("test %v: create: %v", ti.name, err)
			return
		}
		wcleanup := func() {
			if err := w.Close(); err != nil {
				t.Errorf("test %v: error closing file: %v", ti.name, err)
			}
		}

		hIn, hOut := sha1.New(), sha1.New()
		hw := io.MultiWriter(w, hIn)

		err = testutil.WriteRandFile(hw, testutil.KB)
		if err != nil {
			t.Errorf("test %v: write: %v", ti.name, err)
			wcleanup()
			return
		}

		wcleanup()

		// We do this a few times in case someone is using a single
		// bytes.Buffer or some other non-idempotent mechanism for file reads.
		for j := 0; j < readTimes; j++ {
			r, err := fs.Open(fname)
			if err != nil {
				t.Errorf("test %v: open: %v", ti.name, err)
				return
			}
			defer r.Close()

			hOut.Reset()

			n, err := io.Copy(hOut, r)
			if n == 0 {
				t.Errorf("test %v: copy wrote 0 bytes", ti.name)
				return
			}
			if err != nil {
				t.Errorf("test %v: copy: %v", ti.name, err)
				return
			}

			if !bytes.Equal(hIn.Sum(nil), hOut.Sum(nil)) {
				t.Errorf("test %v: hashes did not match", ti.name)
			}
		}
	}

	for i, ti := range tests {
		testBody(i, ti)
	}
}

// TestTouchStat checks that creating a file with 0 length works properly.
func TestTouchStat(t *testing.T) {
	// In this particular case, the file name and the path relative to the fs
	// root are the same.
	want := util.FileInfo{I_Name: "foo"}

	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			fmt.Println(fs, cleanup, err)
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		f, err := fs.Create(want.Name())
		if err != nil {
			t.Errorf("test %v: file creation: %v", ti.name, err)
			return
		}

		f.Close()

		got, err := fs.Stat(want.Name())
		if err != nil {
			t.Errorf("test %v: file stat: %v", ti.name, err)
			return
		}

		checkFileInfo(t, ti.name, got, want)
	}

	for i, ti := range tests {
		testBody(i, ti)
	}
}

// TestTouchRemove checks that remove deletes causes files to no longer shows
// up in Stat or GetFiles.
func TestTouchRemove(t *testing.T) {
	fname := "test"
	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		f, err := fs.Create(fname)
		if err != nil {
			t.Errorf("test %v: create: %v", ti.name, err)
			return
		}
		f.Close()

		err = fs.Remove(fname)
		if err != nil {
			t.Errorf("test %v: remove: %v", ti.name, err)
		}

		_, err = fs.Stat(fname)
		if err == nil {
			t.Errorf("test %v: stat found '%v', but the file should have been deleted", ti.name, fname)
		}
	}

	for i, ti := range tests {
		testBody(i, ti)
	}
}

// TestGetFiles creates some files in a test directory and verifies that
// GetFiles lists them when called. The files returned do *not* have to be in
// any particular order.
func TestGetFiles(t *testing.T) {
	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		newInfo := func(name string) util.FileInfo {
			return util.FileInfo{
				I_Name:  name,
				I_Size:  0,
				I_IsDir: false,
			}
		}

		wantSlice := []util.FileInfo{
			newInfo("file1"),
			newInfo("file2"),
			newInfo("file3"),
			newInfo("file4"),
		}

		wantMap := make(map[string]util.FileInfo)
		for _, info := range wantSlice {
			wantMap[info.Name()] = info
		}

		for _, info := range wantSlice {
			f, err := fs.Create(info.Name())
			if err != nil {
				t.Errorf("test %v: test initialization: %v", ti.name, err)
				continue
			}

			f.Close()
		}

		infos, err := fs.GetFiles("/")
		if err != nil {
			t.Errorf("test %v: GetFiles: %v", ti.name, err)
			return
		}

		if ilen, flen := len(infos), len(wantSlice); ilen != flen {
			t.Errorf("test %v: GetFiles returned list of length %v, want length %v", ti.name, ilen, flen)
			return
		}

		checked := make(map[string]bool)
		for _, got := range infos {
			if want, ok := wantMap[got.Name()]; ok {
				checked[got.Name()] = true
				checkFileInfo(t, ti.name, got, want)
			} else {
				t.Errorf("GetFiles returned a file that shouldn't exist: %v", got.Name())
			}
		}

		for _, info := range wantSlice {
			if !checked[info.Name()] {
				t.Errorf("file %v should exist but was not listed by GetFiles", info.Name())
			}
		}
	}

	for i, ti := range tests {
		testBody(i, ti)
	}
}

// TestMkdir creates a directory and then stats it. If the os.FileInfo that is
// returned does not indicate IsDir() == true, the test fails.
func TestMkdir(t *testing.T) {
	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		dname := "testdir"

		err = fs.Mkdir(dname)
		if err != nil {
			t.Errorf("test %v: mkdir error: %v", ti.name, err)
			return
		}

		info, err := fs.Stat(dname)
		if err != nil {
			t.Errorf("test %v: stat error: %v", ti.name, err)
			return
		}

		if !info.IsDir() {
			t.Errorf("test %v: stat info indicates this is a regular file, not a directory", ti.name)
		}
	}

	for i, ti := range tests {
		testBody(i, ti)
	}
}
