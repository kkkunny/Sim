//go:build !lex && !parse && !analyse && !codegenir && !optimize && !codegenasm

package main

import (
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/compiler/interpret"
	"github.com/kkkunny/Sim/compiler/util"
)

func TestDebug(t *testing.T) {
	llvm.EnablePrettyStackTrace()
	target := stlerror.MustWith(util.GetLLVMTarget())
	path := stlerror.MustWith(stlos.NewFilePath("examples/main.sim").Abs())
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, path))
	ret := stlerror.MustWith(interpret.Interpret(module))
	stltest.AssertEq(t, ret, 0)
}
