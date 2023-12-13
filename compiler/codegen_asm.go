//go:build codegenasm

package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_asm"
	"github.com/kkkunny/Sim/mir"
)

func main() {
	stlerror.Must(llvm.InitializeNativeAsmPrinter())

	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	reader := stlerror.MustWith(codegen_asm.CodegenAsm(mir.DefaultTarget(), path))
	defer reader.Close()
	stlerror.MustWith(io.Copy(os.Stdout, reader))
}
