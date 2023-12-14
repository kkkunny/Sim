package main

import (
	"path/filepath"
	"testing"

	stlerror "github.com/kkkunny/stl/error"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
	"github.com/kkkunny/Sim/output/jit"
)

func TestDebug(t *testing.T) {
	path := stlerror.MustWith(filepath.Abs("example/main.sim"))
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), path))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	llvmModule := outputer.Module()
	stlerror.Must(llvmModule.Verify())
	ret := stlerror.MustWith(jit.RunJit(llvmModule))
	stltest.AssertEq(t, ret, 0)
}