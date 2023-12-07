//go:build !lex && !parse && !analyse && !codegenir && !codegenasm

package main

import (
	"os"
	"path/filepath"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	_ "github.com/kkkunny/Sim/config"
	"github.com/kkkunny/Sim/output/jit"
)

func main() {
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	module := stlerror.MustWith(codegen_ir.CodegenIr(path))
	os.Exit(int(stlerror.MustWith(jit.RunJit(module))))
}
