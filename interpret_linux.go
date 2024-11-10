//go:build !lex && !parse && !analyse && !codegenir && !codegenasm && linux

package main

import (
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/compiler/interpret"
)

func main() {
	target := stlerror.MustWith(llvm.NativeTarget())
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, stlos.NewFilePath(os.Args[1])))
	ret := stlerror.MustWith(interpret.Interpret(module))
	os.Exit(int(ret))
}
