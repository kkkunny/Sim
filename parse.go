//go:build test && parse

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/src/compiler/parse"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/util"
)

func main() {
	pkgs := util.MustValue(parse.Parse(stlos.Path(os.Args[1])))
	for _, pkg := range pkgs {
		fmt.Printf("[%d](%s)\n", pkg.Priority, pkg.Path)
	}
}
