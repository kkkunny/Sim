//go:build analyze

package main

import (
	"os"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/parse"
	simerr "github.com/kkkunny/Sim/compiler/report"
)

func main() {
	reporter := simerr.NewReporter()
	ast := stlerr.MustWith(parse.Parse(os.Args[1], reporter))
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	node := analyze.NewAnalyzer(reporter).Analyze(ast)
	hir.Print(os.Stdout, node)
}
