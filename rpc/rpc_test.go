package rpc

import (
	"bytes"
	"io"
	"log"
	"sync"
	"testing"

	anet "aaw/net"
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

func TestNormalRPC(t *testing.T) {
	const (
		arg1 int = 1
		arg2 int = 2
		want int = 3
	)

	cli, _ := newTestCliSrv(t, &testServer{})

	methodName := serverPrefix + ".Add"
	var got int
	if err := cli.Call(methodName, arg1, arg2, &got); err != nil {
		t.Errorf("call failed: %v", err)
	}

	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
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

func BenchmarkTCPClientCreate(b *testing.B) {
	const hostport = "localhost:8080"
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

func addRPCBench(b *testing.B, d, e, want int, cli *Client) {
	var got int
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err := cli.Call(serverPrefix+".Add", d, e, &got)
		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}

		if got != want {
			b.Fatal(serverPrefix + ".Add not working properly")
		}
	}
}

func BenchmarkTCPNormRPC(b *testing.B) {
	const hostport = "localhost:8080"

	var (
		d    int = 1
		e    int = 1
		want int = d + e
	)

	srv := NewServer()
	srv.Register(serverPrefix, &testServer{})
	defer srv.Close()
	srv.TCPListen(hostport)

	cli, err := NewClient(anet.TCPDialer(hostport))
	if err != nil {
		b.Fatal(err)
	}

	addRPCBench(b, d, e, want, cli)
}

func BenchmarkPipeNormRPC(b *testing.B) {
	var (
		d    int = 1
		e    int = 1
		want int = d + e
	)

	cli, srv, err := NewPipeCliSrv(serverPrefix, &testServer{})
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Close()

	addRPCBench(b, d, e, want, cli)
}
