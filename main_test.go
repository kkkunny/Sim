package main

import (
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/compiler/codegen_ir"

	"github.com/kkkunny/Sim/compiler/interpret"
)

func TestDebug(t *testing.T) {
	stlerror.Must(llvm.InitializeNativeTarget())
	module := stlerror.MustWith(codegen_ir.CodegenIr(stlerror.MustWith(llvm.NativeTarget()), stlos.NewFilePath("examples/main.sim")))
	ret := stlerror.MustWith(interpret.Interpret(module))
	stltest.AssertEq(t, ret, 0)
}
