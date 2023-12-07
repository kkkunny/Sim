package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/output/jit"
)

func TestDebug(t *testing.T) {
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	path := stlerror.MustWith(filepath.Abs("example/main.sim"))
	module := stlerror.MustWith(codegen_ir.CodegenIr(path))
	os.Exit(int(stlerror.MustWith(jit.RunJit(module))))
}
