package metastore

import (
	"fmt"
	"os"
	"path"
	"testing"
)

type metaDBTest struct {
	name  string
	setup func() (MetaStore, func(), error)
}

var tests = []metaDBTest{
	{"inmem", func() (MetaStore, func(), error) {
		return NewInMem(), func() {}, nil
	}},
	{"sqlite", func() (MetaStore, func(), error) {
		dir := path.Join(os.TempDir(), "metadbtest")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("error creating tmp dir %s: %v", dir, err)
		}

		cleanup := func() { os.RemoveAll(dir) }

		fpath, err := CreateMetaDB(dir)
		if err != nil {
			return nil, cleanup, fmt.Errorf("error setting up metadb: %v", err)
		}

		m, err := NewMetaDB(fpath)
		return m, cleanup, err
	}},
}

func TestMetaStoreEndToEnd(t *testing.T) {
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

		if err := m.Remove(key); err != nil {
			t.Errorf("remove error: %v", err)
		}

		// Make sure remove was successful.
		if _, err := m.IsBig(key); err == nil {
			t.Errorf("item exists, after it was removed")
		}
	}

	for _, test := range tests {
		testFunc(test)
	}
}

func TestMetaDBCheckerNoDB(t *testing.T) {
	dir := path.Join(os.TempDir(), "TestMetaDBCheckerNoDB")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Errorf("error creating tmp dir %s: %v", dir, err)
	}
	defer os.RemoveAll(dir)

	if err := CheckMetaDB(path.Join(dir, sqliteDBName)); err == nil {
		t.Errorf("should get 'unable top open file' error")
	}
}

func TestMetaDBCheckerTableOk(t *testing.T) {
	dir := path.Join(os.TempDir(), "TestMetaDBCheckerTableOk")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Errorf("error creating tmp dir %s: %v", dir, err)
	}
	defer os.RemoveAll(dir)

	_, err := CreateMetaDB(dir)
	if err != nil {
		t.Errorf("error in CreateMetaDB: %v", err)
	}

	if err := CheckMetaDB(dir); err != nil {
		t.Errorf("error in CheckMetaDB: %v", err)
	}
}
