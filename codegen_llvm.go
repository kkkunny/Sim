//go:build codegenllvm

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
)

func main() {
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath(os.Args[1])))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	fmt.Println(outputer.Module())
}
