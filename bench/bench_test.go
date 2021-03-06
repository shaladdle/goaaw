package bench

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"
)

func TestBWBench(t *testing.T) {
	const hostport = "localhost:8000"

	type bwresult struct {
		result BWBenchResult
		err    error
	}

	srvResult := make(chan bwresult)
	go func() {
		result, err := StartServerBW(hostport, mywriter{})
		srvResult <- bwresult{result, err}
	}()

	time.Sleep(100 * time.Millisecond)

	cliResult := make(chan bwresult)
	go func() {
		result, err := StartClientBW(hostport)
		cliResult <- bwresult{result, err}
	}()

	r1 := <-cliResult
	if r1.err != nil {
		t.Fatal(r1.err)
	}
	fmt.Println("Client results:", r1.result)

	r2 := <-srvResult
	if r2.err != nil {
		t.Fatal(r2.err)
	}
	fmt.Println("Server results:", r2.result)
}

func TestDiskBench(t *testing.T) {
	filename := path.Join(os.TempDir(), "tmp.dat")

	if result, err := StartDiskBench(filename); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("Disk Result:", result)
	}
}
