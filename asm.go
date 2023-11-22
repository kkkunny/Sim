//go:build asm

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/util"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	target := stlerror.MustWith(llvm.NativeTarget())

	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	asts := stlerror.MustWith(parse.ParseFile(path))
	generator := codegen.New(target, analyse.New(path, asts, target))
	module := generator.Codegen()
	stlerror.Must(module.Verify())

	fmt.Println(module)
	stlerror.Must(target.WriteASMToFile(module, "main.s", llvm.CodeOptLevelNone, llvm.RelocModePIC, llvm.CodeModelDefault))
}
