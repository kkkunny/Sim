//go:build parse

package main

import (
	"os"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
)

func main() {
	file := stlerr.MustWith(os.Open(os.Args[1]))
	defer file.Close()
	reporter := report.NewReporter()
	fileAst := parse.New(lex.New(reader.NewFile(os.Args[1], file)), reporter).Parse()
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}
	ast.Print(os.Stdout, fileAst)
}
