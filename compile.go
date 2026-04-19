//go:build compile

package main

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/compile"
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
	ast := parse.New(lexer, reporter).Parse()
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	node := analyze.NewAnalyzer(reporter).Analyze(ast)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	builder := codegen.New().Generate(node)
	err := compile.NewCompiler(builder).Compile()
	if err != nil {
		panic(err)
	}
}
