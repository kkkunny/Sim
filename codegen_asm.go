//go:build codegenasm

package main

import (
	"io"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_asm"
	"github.com/kkkunny/Sim/mir"
)

func main() {
	stlerror.Must(llvm.InitializeNativeAsmPrinter())

	reader := stlerror.MustWith(codegen_asm.CodegenAsm(mir.DefaultTarget(), stlos.NewFilePath(os.Args[1])))
	defer reader.Close()
	stlerror.MustWith(io.Copy(os.Stdout, reader))
}
