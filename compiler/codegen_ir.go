//go:build codegenir

package main

import (
	"fmt"
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
)

func main() {
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	module := stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), path))
	fmt.Println(module)
}
