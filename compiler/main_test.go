package main

import (
	"testing"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
	"github.com/kkkunny/Sim/output/jit"
)

func TestDebug(t *testing.T) {
	mirModule := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath("example/main.sim")))
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	llvmModule := outputer.Module()
	stlerror.Must(llvmModule.Verify())
	ret := stlerror.MustWith(jit.RunJit(llvmModule))
	stltest.AssertEq(t, ret, 0)
}