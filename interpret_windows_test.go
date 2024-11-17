//go:build !lex && !parse && !analyse && !codegenir && !codegenasm && windows

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
	llvm.EnablePrettyStackTrace()
	stlerror.Must(llvm.InitializeTargetInfo(llvm.X86))
	stlerror.Must(llvm.InitializeTarget(llvm.X86))
	stlerror.Must(llvm.InitializeTargetMC(llvm.X86))
	target := stlerror.MustWith(llvm.NewTargetFromTriple("x86_64-pc-windows-msvc"))
	path := stlerror.MustWith(stlos.NewFilePath("examples/main.sim").Abs())
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, path))
	ret := stlerror.MustWith(interpret.Interpret(module))
	stltest.AssertEq(t, ret, 0)
}
