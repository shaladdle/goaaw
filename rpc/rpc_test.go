package rpc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
)

type TestSrv struct {
	data map[string][]byte
	*sync.Mutex
}

func newTestSrv() *TestSrv {
	return &TestSrv{
		data:  make(map[string][]byte),
		Mutex: &sync.Mutex{},
	}
}

func (t *TestSrv) PutFile(key string, length int64) (io.Writer, error) {
	r, w := io.Pipe()

    // Don't want to just give them the buffer because we want to atomically
    // assign the newly put buffer to the map
	go func() {
		b := &bytes.Buffer{}

		_, err := io.CopyN(b, r, length)
		if err == nil {
			t.Lock()

			t.data[key] = b.Bytes()

			t.Unlock()
		}
	}()

	return w, nil
}

func (t *TestSrv) GetFile(key string) (io.Reader, error) {
	t.Lock()
	defer t.Unlock()

	if b, ok := t.data[key]; ok {
		return bytes.NewReader(b), nil
	}

	return nil, fmt.Errorf("file does not exist")
}

func (t *TestSrv) GetSize(key string, size *int) error {
	t.Lock()
	defer t.Unlock()

    if b, ok := t.data[key]; ok {
        *size = len(b)
    } else {
        return fmt.Errorf("file does not exist")
    }

	return nil
}

type fileSpec struct {
	key, content string
}

func fillTestSrv(ts *TestSrv) ([]fileSpec, error) {
	files := []fileSpec{
		{"f1", "this is the contents of file 1"},
		{"f2", "this is file 2"},
		{"wat", "watwatwatwat"},
	}

	contents := make([]string, len(files))
	for i, file := range files {
		b := bytes.NewBufferString(file.content)
		w, err := ts.PutFile(file.key, int64(len(file.content)))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, b)
		if err != nil {
			return nil, err
		}

		contents[i] = file.content
	}

	return files, nil
}

// TestTestSrv tests the mock server object.
func TestTestSrv(t *testing.T) {
	ts := newTestSrv()
	files, err := fillTestSrv(ts)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		r, err := ts.GetFile(file.key)
		if err != nil {
			t.Error("error:", err)
			continue
		}

		b := &bytes.Buffer{}
		_, err = io.Copy(b, r)
		if err != nil {
			t.Error("error:", err)
			continue
		}
	}
}

func TestServer(t *testing.T) {
	addr := ":8000"
	ts := newTestSrv()
	s := NewServer()
	err := s.Register("TestSrv", ts)
	if err != nil {
		t.Fatal(err)
	}

	files, err := fillTestSrv(ts)
	if err != nil {
		t.Fatal(err)
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go s.Accept(lis)

	c, err := NewClient(TcpDialer(addr))
	if err != nil {
		t.Fatal(err)
	}

	getFile := func(f fileSpec) error {
		wantStr, wantSize := f.content, len(f.content)
		var gotSize int
		err = c.Call("TestSrv.GetSize", f.key, &gotSize)
		if err != nil {
            return fmt.Errorf("GetSize: %v", err)
		}

		if gotSize != wantSize {
			return fmt.Errorf("incorrect size, got %v, want %v", gotSize, wantSize)
		}

        out, err := c.CallRead("TestSrv.GetFile", f.key)
		if err != nil {
            return fmt.Errorf("GetFile: %v", err)
		}

		b := &bytes.Buffer{}
		_, err = io.Copy(b, out)
		if err != nil {
            return fmt.Errorf("io.Copy: %v", err)
		}

		gotStr := b.String()
		if gotStr != wantStr {
			return fmt.Errorf("incorrect file, got %v, \"want %v\"", gotStr, wantStr)
		}

		return nil
	}

	errors := make(chan error)
	for _, file := range files {
		go func(file_ fileSpec) {
			errors <- getFile(file_)
		}(file)
	}

	for i := 0; i < len(files); i++ {
		if err := <-errors; err != nil {
			t.Error("Error:", err)
		}
	}
}

func TestChanListener(t *testing.T) {
	cl := make(chanListener)

	conn := &net.TCPConn{}
	go func() {
		cl <- conn
	}()

	rconn, err := cl.Accept()
	if err != nil {
		t.Fatal(err)
	}

	if rconn != conn {
		t.Fatal("conns don't match")
	}
}
