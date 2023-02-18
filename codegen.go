//go:build test && codegen

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/Sim/src/compiler/codegen"
	"github.com/kkkunny/Sim/src/compiler/parse"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/util"
)

func main() {
	ast := util.MustValue(parse.Parse(stlos.Path(os.Args[1])))
	hirs := util.MustValue(analyse.NewAnalyser().Analyse(ast))
	module := codegen.NewCodeGenerator().Codegen(*hirs)
	fmt.Println(module)
}
