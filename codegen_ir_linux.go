//go:build codegenir && linux

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir/v2"
)

func main() {
	target := stlerror.MustWith(llvm.NativeTarget())
	module := stlerror.MustWith(codegen_ir.CodegenIr(target, stlos.NewFilePath(os.Args[1])))
	fmt.Println(module)
}
