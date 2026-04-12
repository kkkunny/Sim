//go:build !lex && !parse && !analyse && !codegenir && !optimize && !codegenasm

package main

import (
	"os"
	"path/filepath"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/compiler/interpret"
	"github.com/kkkunny/Sim/compiler/util"
)

func main() {
	llvm.EnablePrettyStackTrace()
	target := stlerror.MustWith(util.GetLLVMTarget())
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, path))
	ret := stlerror.MustWith(interpret.Interpret(module))
	os.Exit(int(ret))
}
