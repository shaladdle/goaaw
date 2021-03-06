package main

import (
	"flag"
	"fmt"

	"github.com/shaladdle/goaaw/bench"
)

var addr = flag.String("addr", "localhost:8000", "")

func main() {
	flag.Parse()

	result, err := bench.StartClientBW(*addr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Client:", result)
}
