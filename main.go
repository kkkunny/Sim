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
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	target := stlerror.MustWith(llvm.NativeTarget())
	mean.Usize.Bits = target.PointerSize() * 8
	mean.Isize.Bits = mean.Usize.Bits
	source, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer source.Close()
	generator := codegen.New(target, analyse.New(parse.New(lex.New(r))))
	module := generator.Codegen()
	engine := stlerror.MustWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	mainFn := engine.GetFunction("main")
	if mainFn == nil {
		panic("can not fond the main function")
	}
	ret := uint8(engine.RunFunction(*mainFn).Integer(false))
	os.Exit(int(ret))
}
