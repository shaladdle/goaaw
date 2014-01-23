package main

import (
    "fmt"
    "os"

    "github.com/shaladdle/goaaw/dedup/simple"
)

func main() {
    d := simple.NewDeduper(4096)

    if len(os.Args) != 2 {
        fmt.Println("usage: dedup [filename]")
        return
    }

    list, err := d.ComputeBlockList(os.Args[1])
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    for _, blkInfo := range list {
        fmt.Printf("%0x: {%d, %d}\n", blkInfo.Hash, blkInfo.Pos, blkInfo.Size)
    }
}
