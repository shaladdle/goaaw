package bench

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"time"
)

type mywriter struct{}

func (mywriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// 1MB
const (
	mb              int64 = 1024 * 1024
	gb              int64 = 1024 * mb
	bwBenchNumBytes       = 100 * mb
)

type BWBenchResult struct {
	Start    time.Time
	Duration time.Duration
	NumBytes int64
}

func (r BWBenchResult) String() string {
	return fmt.Sprintf("%.2f MB/s", float64(r.NumBytes/mb)/r.Duration.Seconds())
}

// StartClientBW dials
func StartClientBW(hostport string) (BWBenchResult, error) {
	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		return BWBenchResult{}, err
	}
	defer conn.Close()

	b := &bytes.Buffer{}
	if _, err := io.CopyN(b, rand.Reader, bwBenchNumBytes); err != nil {
		return BWBenchResult{}, err
	}

	start := time.Now()
	if _, err := io.CopyN(conn, b, bwBenchNumBytes); err != nil {
		return BWBenchResult{}, err
	}
	duration := time.Since(start)

	return BWBenchResult{start, duration, bwBenchNumBytes}, nil
}

// StartServerBW listens on the specified hostport for a client connection,
// reads the data coming through, and measures how long it took to receive.
func StartServerBW(hostport string) (BWBenchResult, error) {
	l, err := net.Listen("tcp", hostport)
	if err != nil {
		return BWBenchResult{}, err
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return BWBenchResult{}, err
	}
	defer conn.Close()

	start := time.Now()
	_, err = io.CopyN(mywriter{}, conn, bwBenchNumBytes)
	if err != nil {
		return BWBenchResult{}, err
	}
	duration := time.Since(start)

	return BWBenchResult{start, duration, bwBenchNumBytes}, nil
}
