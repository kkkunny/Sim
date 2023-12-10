package main

import (
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
	"github.com/kkkunny/Sim/output/jit"
)

func main() {
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), path))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	llvmModule := outputer.Module()
	stlerror.Must(llvmModule.Verify())
	ret := stlerror.MustWith(jit.RunJit(llvmModule))
	os.Exit(int(ret))
}