//go:build parse

package main

import (
	"os"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/parse"
	simerr "github.com/kkkunny/Sim/compiler/report"
)

func main() {
	reporter := simerr.NewReporter()
	node := stlerr.MustWith(parse.Parse(os.Args[1], reporter))

	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	ast.Print(os.Stdout, node)
}
