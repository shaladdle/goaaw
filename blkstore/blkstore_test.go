package blkstore

import (
	"testing"

	"aaw/fs/remote"
	anet "aaw/net"
	"aaw/testutil"
)

type testCase struct {
	name  string
	setup func(t *testing.T) (bs BlkStore, cleanup func())
}

var tests = []testCase{
	{"inmem", func(t *testing.T) (BlkStore, func()) {
		return NewMemStore(), func() {}
	}},
	{"disk", func(t *testing.T) (BlkStore, func()) {
		te := testutil.NewTestEnv("disk", t)
		return NewDiskStore(te.Root()), func() { te.Teardown() }
	}},
	{"remote", func(t *testing.T) (BlkStore, func()) {
		const hostport = "localhost:9000"
		te := testutil.NewTestEnv("remote", t)

		srv, err := remote.NewTCPServer(te.Root(), hostport)
		if err != nil {
			t.Errorf("setup error: %v", err)
		}

		bs, err := NewRemoteStore(anet.TCPDialer(hostport))
		if err != nil {
			t.Errorf("setup error: %v", err)
		}

		return bs, func() {
			srv.Close()
			te.Teardown()
		}
	}},
}

func testPutGet(t *testing.T, test testCase) {
	bs, cleanup := test.setup(t)
	defer cleanup()

	m := map[string]string{
		"a": "woeifjowiejf",
		"c": "oij11290909",
		"l": "next string",
		"e": "0-f1-02j-921j4",
		"d": ";;;12;31",
	}

	for k, want := range m {
		if err := bs.Put(k, []byte(want)); err != nil {
			t.Errorf("%v: put error for key %v: %v", test.name, k, err)
		}

		if got, err := bs.Get(k); err != nil {
			t.Errorf("%v: get error for key %v: %v", test.name, k, err)
		} else if string(got) != want {
			t.Errorf("%v: got %v, want %v", test.name, string(got), want)
		}
	}
}

func TestPutGet(t *testing.T) {
	for _, test := range tests {
		testPutGet(t, test)
	}
}

func testDelete(t *testing.T, test testCase) {
	bs, cleanup := test.setup(t)
	defer cleanup()

	const (
		key  = "key"
		want = "value"
	)

	if _, err := bs.Get(key); err == nil {
		t.Errorf("%v: get didn't return an error, but it should", test.name)
	}

	if err := bs.Put(key, []byte(want)); err != nil {
		t.Errorf("%v: put error: %v", test.name, err)
	}

	if _, err := bs.Get(key); err != nil {
		t.Errorf("%v: get error after putting: %v", test.name, err)
	}

	if err := bs.Delete(key); err != nil {
		t.Errorf("%v: get error after putting: %v", test.name, err)
	}

	if _, err := bs.Get(key); err == nil {
		t.Errorf("%v: get didn't return an error, but it should", test.name)
	}
}

func TestDelete(t *testing.T) {
	for _, test := range tests {
		testDelete(t, test)
	}
}
