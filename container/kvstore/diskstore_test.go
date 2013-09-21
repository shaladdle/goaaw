package kvstore

import (
	"os"
	"path"
	"testing"
)

func TestDiskstore(t *testing.T) {
	fpath := path.Join(os.TempDir(), "diskstore-tmp.kvstore")
	defer os.RemoveAll(fpath)
	s, e := NewGobKVStore(fpath, "", "")
	if e != nil {
		t.Fatal(e)
	}

	const (
		wantKey = "hi"
		wantVal = "this is the data"
	)

	if err := s.Put(wantKey, wantVal); err != nil {
		t.Errorf("error putting: %v", err)
	}

	if gotVal, err := s.Get(wantKey); err != nil {
		t.Errorf("error getting %v", err)
	} else if gotVal != wantVal {
		t.Errorf("got %v, want %v", gotVal, wantVal)
	}

	s1, e := NewGobKVStore(fpath, "", "")
	if e != nil {
		t.Fatal(e)
	}

	if gotVal, err := s1.Get(wantKey); err != nil {
		t.Errorf("error getting %v", err)
	} else if gotVal != wantVal {
		t.Errorf("got %v, want %v", gotVal, wantVal)
	}
}
