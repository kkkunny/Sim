//go:build codegenasm && windows

package main

import (
	"io"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_asm"
)

func main() {
	stlerror.Must(llvm.InitializeTargetInfo(llvm.X86))
	stlerror.Must(llvm.InitializeTarget(llvm.X86))
	stlerror.Must(llvm.InitializeTargetMC(llvm.X86))
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	target := stlerror.MustWith(llvm.NewTargetFromTriple("x86_64-pc-windows-msvc"))
	path := stlerror.MustWith(stlos.NewFilePath(os.Args[1]).Abs())
	reader := stlerror.MustWith(codegen_asm.CodegenAsm(target, path))
	defer reader.Close()
	stlerror.MustWith(io.Copy(os.Stdout, reader))
}
