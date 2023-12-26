//go:build !lex && !parse && !analyse && !codegenir && !codegenllvm && !codegenasm
package main

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/interpret"
	"github.com/kkkunny/Sim/mir"
)

func main() {
	module := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath(os.Args[1])))
	ret := stlerror.MustWith(interpret.Interpret(module))
	os.Exit(int(ret))
}