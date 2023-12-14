//go:build codegenllvm

package main

import (
	"fmt"
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
)

func main() {
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), path))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	fmt.Println(outputer.Module())
}
