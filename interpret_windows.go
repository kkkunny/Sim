//go:build !lex && !parse && !analyse && !codegenir && !codegenasm && windows

package main

import (
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"

	"github.com/kkkunny/Sim/compiler/interpret"
)

func main() {
	stlerror.Must(llvm.InitializeTargetInfo(llvm.X86))
	stlerror.Must(llvm.InitializeTarget(llvm.X86))
	stlerror.Must(llvm.InitializeTargetMC(llvm.X86))
	target := stlerror.MustWith(llvm.NewTargetFromTriple("x86_64-pc-windows-msvc"))
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, stlos.NewFilePath(os.Args[1])))
	ret := stlerror.MustWith(interpret.Interpret(module))
	os.Exit(int(ret))
}
