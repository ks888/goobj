package main

import (
	"fmt"
	"os"

	"github.com/ks888/goobj"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [go object file]\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer f.Close()

	file, err := goobj.Parse(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse goobj file: %v\n", err)
		os.Exit(1)
	}

	goobj.PrintSymbols(file)
}
