//go:build compile

package main

import (
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/compile"
)

func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}
	builder := codegen.New().Generate(pkg)
	err = compile.NewCompiler(builder).Compile()
	if err != nil {
		panic(err)
	}
}
