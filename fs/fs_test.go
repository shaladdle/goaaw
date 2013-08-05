package fs

// TODO: Get in memory file system working

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"testing"

	"aaw/fs/remote"
	"aaw/fs/std"
	"aaw/testutil"
)

type testInfo struct {
	name  string
	setup func(*testing.T) (FileSystem, func(), error)
}

var tests = []testInfo{
	{"std", func(t *testing.T) (FileSystem, func(), error) {
		te := testutil.NewTestEnv("testcase-stdfs", t)
		return std.New(te.Root()), func() { te.Teardown() }, nil
	}},
	{"remote", func(t *testing.T) (FileSystem, func(), error) {
		te := testutil.NewTestEnv("testcase-remotefs", t)

		cli, srv, err := remote.NewPipeCliSrv(te.Root())
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
			err := w.Close()
			if err != nil {
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
	fname := "foo"

	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			fmt.Println(fs, cleanup, err)
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		f, err := fs.Create(fname)
		if err != nil {
			t.Errorf("test %v: file creation: %v", ti.name, err)
			return
		}

		f.Close()

		info, err := fs.Stat(fname)
		if err != nil {
			t.Errorf("test %v: file stat: %v", ti.name, err)
			return
		}

		if info.Name() != fname {
			t.Errorf("test %v: file name mismatch, got %v, want %v", ti.name, info.Name(), fname)
		}

		if info.Size() != 0 {
			t.Errorf("test %v: file size should be 0", ti.name, info.Name(), fname)
		}
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
// GetFiles lists them when called.
//
// TODO: Test all fields of the FileInfo interface
func TestGetFiles(t *testing.T) {
	testBody := func(i int, ti testInfo) {
		fs, cleanup, err := ti.setup(t)
		defer cleanup()
		if err != nil {
			t.Errorf("test %v: test initialization: %v", ti.name, err)
			return
		}

		fnames := []string{"file1", "file2", "file3", "file4"}

		for _, fname := range fnames {
			f, err := fs.Create(fname)
			if err != nil {
				t.Errorf("test %v: test initialization: %v", ti.name, err)
				continue
			}

			err = testutil.WriteRandFile(f, 0)
			if err != nil {
				t.Errorf("test %v: writing random file '%v': %v", ti.name, fname, err)
				f.Close()
				continue
			}

			f.Close()
		}

		infos, err := fs.GetFiles("/")
		if err != nil {
			t.Errorf("test %v: GetFiles: %v", ti.name, err)
			return
		}

		if ilen, flen := len(infos), len(fnames); ilen != flen {
			t.Errorf("test %v: GetFiles returned list of length %v, want length %v", ti.name, ilen, flen)
			return
		}

		for i, info := range infos {
			if info.Name() != fnames[i] {
				t.Errorf("test %v: name mismatch, got %v, want %v", ti.name, info.Name(), fnames[i])
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
