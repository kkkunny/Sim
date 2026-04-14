package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sim <source_file>")
		os.Exit(1)
	}

	sourcePath := os.Args[1]
	outputPath := "a.out"
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}

	err := compiler.Compile(sourcePath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}
}
