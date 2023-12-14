//go:build codegenir

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
)

func main() {
	module := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), stlos.NewFilePath(os.Args[1])))
	fmt.Println(module)
}
