//go:build codegenir

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
)

func main() {
	stlerror.Must(llvm.InitializeNativeTarget())
	module := stlerror.MustWith(codegen_ir.CodegenIr(stlerror.MustWith(llvm.NativeTarget()), stlos.NewFilePath(os.Args[1])))
	fmt.Println(module)
}
