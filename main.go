//go:build !lex && !parse && !analyse && !codegenir && !codegenasm

package main

import (
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"

	"github.com/kkkunny/Sim/compiler/interpret"
)

type name struct {
	a int32
}

func main() {
	stlerror.Must(llvm.InitializeNativeTarget())
	module := stlerror.MustWith(codegen_ir.CodegenIr(stlerror.MustWith(llvm.NativeTarget()), stlos.NewFilePath(os.Args[1])))
	ret := stlerror.MustWith(interpret.Interpret(module))
	os.Exit(int(ret))
}
