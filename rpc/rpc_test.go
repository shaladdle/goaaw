package rpc

import (
	"bytes"
	"io"
	"log"
	"sync"
	"testing"

	anet "github.com/shaladdle/goaaw/net"
)

// testServer is the type used as the server side of the RPC connection. This
// example requires the use of a mutex, because the Write function waits until
// the connection is closed before committing the data. If the client were to
// make a read call after the connection is closed and before the write data
// was committed, it would seem like the write hadn't happened. This might
// be acceptable in some situations, but it is good to note this limitation.
type testServer struct {
	sync.Mutex
	data []byte
}

func (testServer) RPCNorm_Add(a, b int) int {
	return a + b
}

func (testServer) RPCNorm_Range() (a, b, c int, d string) {
	return 1, 2, 3, "hi there"
}

func (s *testServer) RPCRead_ReadData() (io.Reader, StrError) {
	s.Lock()
	defer s.Unlock()
	if len(s.data) == 0 {
		return nil, StrError("No data")
	}

	return bytes.NewReader(s.data), ErrNil
}

func (s *testServer) RPCWrite_WriteData() (io.WriteCloser, StrError) {
	r, w := io.Pipe()

	s.Lock()
	go func() {
		defer s.Unlock()
		b := &bytes.Buffer{}
		_, err := io.Copy(b, r)
		if err != nil {
			log.Println("error in WriteData:", err)
			return
		}

		s.data = b.Bytes()
	}()

	return w, ErrNil
}

func (s *testServer) GetData() []byte {
	s.Lock()
	defer s.Unlock()

	return s.data
}

const serverPrefix = "Test"

func newTestCliSrv(t *testing.T, s interface{}) (*Client, *Server) {
	pnet := anet.NewPipeNet()
	rpcs := NewServer()
	rpcs.Register(serverPrefix, s)
	go rpcs.Accept(pnet)

	cli, err := NewClient(pnet)
	if err != nil {
		t.Errorf("client creation failed: %v", err)
	}

	return cli, rpcs
}

func testAdd(t *testing.T, cli *Client) {
	const (
		arg1 int = 1
		arg2 int = 2
		want int = 3
	)

	methodName := serverPrefix + ".Add"
	var got int
	if err := cli.Call(methodName, arg1, arg2, &got); err != nil {
		t.Errorf("call failed: %v", err)
	}

	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

type tester interface {
	Errorf(string, ...interface{})
}

func testRange(t tester, cli *Client) {
	wanta, wantb, wantc := 1, 2, 3
	wantd := "hi there"

	var (
		gota, gotb, gotc int
		gotd             string
	)

	methodName := serverPrefix + ".Range"

	switch b := t.(type) {
	case *testing.B:
		b.StartTimer()
	}

	if err := cli.Call(methodName, &gota, &gotb, &gotc, &gotd); err != nil {
		t.Errorf("call failed: %v", err)
	}

	switch b := t.(type) {
	case *testing.B:
		b.StopTimer()
	}

	if gota != wanta {
		t.Errorf("wanta %v, gota %v", wanta, gota)
	}

	if gotb != wantb {
		t.Errorf("wantb %v, gotb %v", wantb, gotb)
	}

	if gotc != wantc {
		t.Errorf("wantc %v, gotc %v", wantc, gotc)
	}

	if gotd != wantd {
		t.Errorf("wantd %v, gotd %v", wantd, gotd)
	}
}

func TestNormalRPC(t *testing.T) {
	cli, _ := newTestCliSrv(t, &testServer{})

	testAdd(t, cli)
	testRange(t, cli)
}

func TestReadRPC(t *testing.T) {
	const want = "streaming data"

	s := &testServer{data: []byte(want)}
	cli, _ := newTestCliSrv(t, s)

	var callErr StrError
	r, err := cli.CallRead(serverPrefix+".ReadData", &callErr)
	if err != nil {
		t.Fatalf("CallRead error: %v", err)
	}

	b := make([]byte, len(want))
	if _, err := r.Read(b); err != nil {
		t.Fatalf("Read error: %v", err)
	}

	got := string(b)
	if got != want {
		t.Errorf("got '%v', want '%v'", got, want)
	}
}

func TestWriteRPC(t *testing.T) {
	want := []byte("streaming data")

	s := &testServer{}
	cli, _ := newTestCliSrv(t, s)

	var callErr StrError
	w, err := cli.CallWrite(serverPrefix+".WriteData", &callErr)
	if err != nil {
		t.Fatalf("CallWrite error: %v", err)
	}

	if _, err := w.Write(want); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	w.Close()

	if got := s.GetData(); !bytes.Equal(got, want) {
		t.Errorf("got '%v', want '%v'", string(got), string(want))
	}
}

func TestReadWriteRPC(t *testing.T) {
	const want = "streaming data"

	s := &testServer{}
	cli, _ := newTestCliSrv(t, s)

	var callErr StrError
	w, err := cli.CallWrite(serverPrefix+".WriteData", &callErr)
	if err != nil {
		t.Fatalf("CallWrite error: %v", err)
	}

	if _, err := w.Write([]byte(want)); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	w.Close()

	r, err := cli.CallRead(serverPrefix+".ReadData", &callErr)
	if err != nil {
		t.Fatalf("CallRead error: %v", err)
	}

	b := make([]byte, len(want))
	if _, err := r.Read(b); err != nil {
		t.Fatalf("Read error: %v", err)
	}

	got := string(b)
	if got != want {
		t.Errorf("got '%v', want '%v'", got, want)
	}
}

// TODO: Add test case to make sure server can handle nil return values on a
// streaming RPC.

const hostport = "localhost:9000"

func BenchmarkTCPClientCreate(b *testing.B) {
	srv := NewServer()
	defer srv.Close()
	srv.TCPListen(hostport)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewClient(anet.TCPDialer(hostport))
	}
	b.StopTimer()
}

func BenchmarkPipeClientCreate(b *testing.B) {
	pnet := anet.NewPipeNet()
	srv := NewServer()
	defer srv.Close()
	go srv.Accept(pnet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewClient(pnet)
	}
	b.StopTimer()
}

func rangeRPCBench(b *testing.B, cli *Client) {
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testRange(b, cli)
	}
}

func BenchmarkTCPNormRPC(b *testing.B) {
	srv := NewServer()
	srv.Register(serverPrefix, &testServer{})
	defer srv.Close()
	srv.TCPListen(hostport)

	cli, err := NewClient(anet.TCPDialer(hostport))
	if err != nil {
		b.Fatal(err)
	}

	rangeRPCBench(b, cli)
}

func BenchmarkPipeNormRPC(b *testing.B) {
	cli, srv, err := NewPipeCliSrv(serverPrefix, &testServer{})
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Close()

	rangeRPCBench(b, cli)
}
