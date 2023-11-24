package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/util"

	"github.com/kkkunny/Sim/analyse"
)

func TestDebug(t *testing.T) {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	target := stlerror.MustWith(llvm.NativeTarget())

	path := stlerror.MustWith(filepath.Abs("example/main.sim"))
	asts := stlerror.MustWith(parse.ParseFile(path))
	generator := codegen.New(target, analyse.New(path, asts, target))
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
