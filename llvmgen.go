//go:build llvmgen

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/llgen"
)

// 迁移期临时驱动：analyze → llgen → 打印 LLVM IR
func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}

	ctx := llgen.NewContext()
	defer ctx.Close()

	for _, depPkg := range pkg.Dependencies {
		gen := llgen.New(ctx, depPkg)
		gen.Generate()
		fmt.Println(gen.Module())
		if err := gen.Close(); err != nil {
			panic(err)
		}
	}

	gen := llgen.New(ctx, pkg)
	gen.Generate()
	fmt.Println(gen.Module())
	if err := gen.Close(); err != nil {
		panic(err)
	}
}
