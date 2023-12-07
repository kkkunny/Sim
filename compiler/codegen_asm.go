//go:build codegenasm

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_asm"
)

func main() {
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	reader := stlerror.MustWith(codegen_asm.CodegenAsm(path, llvm.CodeOptLevelNone))
	defer reader.Close()
	fmt.Println(string(stlerror.MustWith(io.ReadAll(reader))))
}
