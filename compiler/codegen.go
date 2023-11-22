//go:build codegen

package main

import (
	"fmt"
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/util"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())

	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	asts := stlerror.MustWith(parse.ParseFile(path))
	generator := codegen.New(target, analyse.New(path, asts, target))
	module := generator.Codegen()
	fmt.Println(module)
	stlerror.Must(module.Verify())
}