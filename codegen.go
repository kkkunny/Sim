//go:build codegen

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
)

func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}
	ctx := codegen.NewContext()
	for _, depPkg := range pkg.Dependencies {
		codegen.New(ctx, depPkg).Generate()
	}
	builder := codegen.New(ctx, pkg).Generate()
	fmt.Println(builder)
}
