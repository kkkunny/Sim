package test

import (
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	_ "github.com/kkkunny/Sim/config"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func init() {
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
}

func assertRetEq(t *testing.T, code string, expect uint8) {
	target := stlerror.MustWith(llvm.NativeTarget())
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(target, analyse.New(parse.New(lex.New(r)), target)).Codegen()
	engine := stlerror.MustWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	f := engine.GetFunction("main")
	if f == nil {
		panic("can not fond the main function")
	}
	ret := uint8(engine.RunFunction(*f).Integer(true))
	stltest.AssertEq(t, ret, expect)
}

func assertRetEqZero(t *testing.T, code string) {
	assertRetEq(t, code, 0)
}
