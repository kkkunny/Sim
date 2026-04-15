//go:build analyze

package main

import (
	"bytes"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
)

func main() {
	data := stlerror.MustWith(os.ReadFile(os.Args[1]))
	lexer := lex.New(bytes.NewReader(data))
	ast := parse.New(lexer).Parse()
	hir.Print(os.Stdout, analyze.NewAnalyzer().Analyze(ast))
}
