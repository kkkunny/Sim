//go:build codegen

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())

	f, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer f.Close()
	generator := codegen.New(target, analyse.New(parse.New(lex.New(r)), target))
	module := generator.Codegen()
	fmt.Println(module)
	stlerror.Must(module.Verify())
}
