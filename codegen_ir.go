//go:build codegenir

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/compiler/util"
)

func main() {
	llvm.EnablePrettyStackTrace()
	target := stlerror.MustWith(util.GetLLVMTarget())
	path := stlerror.MustWith(stlos.NewFilePath(os.Args[1]).Abs())
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, path))
	fmt.Println(module)
}
