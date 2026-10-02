//go:build codegen

package main

import (
	"fmt"
	"os"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/llgen"
)

func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}
	ctx := llgen.NewContext()
	defer func() {
		stlerr.Must(ctx.Close())
	}()
	for _, depPkg := range pkg.Dependencies {
		g := llgen.New(ctx, depPkg)
		g.Generate()
		stlerr.Must(g.Close())
	}
	g := llgen.New(ctx, pkg)
	model := g.Generate()
	fmt.Println(model)
	stlerr.Must(g.Close())
}
