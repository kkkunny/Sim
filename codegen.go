//go:build codegen

package main

import (
	"bytes"
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
)

func main() {
	data := stlerror.MustWith(os.ReadFile(os.Args[1]))
	lexer := lex.New(bytes.NewReader(data))
	builder := codegen.New().Generate(analyze.NewAnalyzer().Analyze(parse.New(lexer).Parse()))
	fmt.Println(builder)
}
