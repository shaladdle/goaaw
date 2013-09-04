package cloudfs

import (
	"fmt"
	"os"
	"path"
	"testing"
)

type metaDBTest struct {
	name  string
	setup func() (metaStore, func(), error)
}

var tests = []metaDBTest{
	{"inmem", func() (metaStore, func(), error) {
		return newMemMetaStore(), func() {}, nil
	}},
	{"sqlite", func() (metaStore, func(), error) {
		dir := path.Join(os.TempDir(), "metadbtest")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("error creating tmp dir %s: %v", dir, err)
		}

		cleanup := func() { os.RemoveAll(dir) }

		fpath, err := createMetaDb(dir)
		if err != nil {
			return nil, cleanup, fmt.Errorf("error setting up metadb: %v", err)
		}

		m, err := newMetaDB(fpath)
		return m, cleanup, err
	}},
}

func TestMetaDB(t *testing.T) {
	testFunc := func(test metaDBTest) {
		m, cleanup, err := test.setup()
		if err != nil {
			t.Errorf("error creating metadb object: %v", err)
		}
		defer cleanup()

		const (
			key  = "mykey"
			want = true
		)

		// First make sure we get an error on IsBig, since the database should be
		// empty.
		if _, err := m.IsBig(key); err == nil {
			t.Errorf("database was not empty, maybe test environment was not cleaned up")
		}

		// Now actually set it.
		if err := m.SetBig(key, want); err != nil {
			t.Errorf("error calling SetBig: %v", err)
		}

		// Check that we can retrieve it ok and it's what we set it to.
		if got, err := m.IsBig(key); err != nil {
			t.Errorf("error calling IsBig: %v", err)
		} else if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	for _, test := range tests {
		testFunc(test)
	}
}
