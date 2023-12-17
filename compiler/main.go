//go:build !lex && !parse && !analyse && !codegenir && !codegenllvm && !codegenasm
package main

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
	"github.com/kkkunny/Sim/output/jit"
)

func main() {
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath(os.Args[1])))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	llvmModule := outputer.Module()
	stlerror.Must(llvmModule.Verify())
	ret := stlerror.MustWith(jit.RunJit(llvmModule))
	os.Exit(int(ret))
}