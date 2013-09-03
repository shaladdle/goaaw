package dedupfs

import (
	"testing"
)

const (
	TESTDIR = "/tmp/fs/test"
)

func TestNewClose(t *testing.T) {
	fs, err := NewFileSystem(TESTDIR)
	if err != nil {
		t.Fatal("NewFileSystem error:", err)
	}
	err = fs.Close()
	if err != nil {
		t.Fatal("Close error:", err)
	}
}

func TestSimpleIndex(t *testing.T) {
	fs, err := NewFileSystem(TESTDIR)
	if err != nil {
		t.Fatal("NewFileSystem error:", err)
	}

	_, err = fs.Create("hi there")
	if err != nil {
		t.Fatal("Create error:", err)
	}

	err = fs.Close()
	if err != nil {
		t.Fatal("Close error:", err)
	}
}
