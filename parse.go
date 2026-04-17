//go:build parse

package main

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
	"github.com/kkkunny/Sim/compiler/reader"
	simerr "github.com/kkkunny/Sim/compiler/report"
)

func main() {
	file := stlerror.MustWith(os.Open(os.Args[1]))
	defer file.Close()
	lexer := lex.New(reader.NewFile(os.Args[1], file))
	reporter := simerr.NewReporter()
	node := parse.New(lexer, reporter).Parse()

	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	ast.Print(os.Stdout, node)
}
