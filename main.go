//go:build !lex && !parse && !analyse && !codegen

package main

import (
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	_ "github.com/kkkunny/Sim/config"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	target := stlerror.MustWith(llvm.NativeTarget())

	f, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer f.Close()
	generator := codegen.New(target, analyse.New(parse.New(lex.New(r)), target))
	module := generator.Codegen()
	stlerror.Must(module.Verify())
	engine := stlerror.MustWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	mainFn := engine.GetFunction("main")
	if mainFn == nil {
		panic("can not fond the main function")
	}
	ret := uint8(engine.RunFunction(*mainFn).Integer(false))
	os.Exit(int(ret))
}
