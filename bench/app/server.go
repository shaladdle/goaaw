package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaladdle/goaaw/bench"
)

var (
	addr  = flag.String("addr", "localhost:8000", "")
	fname = flag.String("fname", "", "")
)

func main() {
	flag.Parse()

	var (
		result bench.BWBenchResult
		err    error
	)
	switch *fname {
	case "":
		result, err = bench.StartServerBW(*addr, bench.DevNull)
	default:
		f, err := os.Create(*fname)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer f.Close()
		result, err = bench.StartServerBW(*addr, f)
	}
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Server:", result)
}
