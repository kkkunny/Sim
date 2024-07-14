package main

import (
	"testing"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/compiler/codegen_ir"

	"github.com/kkkunny/Sim/compiler/interpret"
	"github.com/kkkunny/Sim/mir"
)

func TestDebug(t *testing.T) {
	module := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath("examples/main.sim")))
	ret := stlerror.MustWith(interpret.Interpret(module))
	stltest.AssertEq(t, ret, 0)
}
