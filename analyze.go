//go:build analyze

package main

import (
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/hir"
)

func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}
	hir.Print(os.Stdout, pkg)
}
