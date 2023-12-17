package test

import (
	"testing"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen_ir"
	_ "github.com/kkkunny/Sim/config"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
	"github.com/kkkunny/Sim/mir/pass/module"
	"github.com/kkkunny/Sim/output/jit"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

func assertRetEq(t *testing.T, code string, expect uint8, skips ...uint) {
	var skip uint
	if len(skips) > 0{
		skip = skips[0]
	}
	path := stlerror.MustWith(stlos.NewFilePath(stlerror.MustWith(util.GetFileName(skip+1))).Abs())
	r := stlerror.MustWith(reader.NewReaderFromString(path, code))
	mirModule := codegen_ir.New(mir.DefaultTarget(), analyse.New(parse.New(lex.New(r)).Parse()).Analyse()).Codegen()
	module.Run(mirModule, module.DeadCodeElimination)
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(mirModule)
	llvmModule := outputer.Module()
	stltest.AssertEq(t, stlerror.MustWith(jit.RunJit(llvmModule)), expect)
}

func assertRetEqZero(t *testing.T, code string) {
	assertRetEq(t, code, 0, 1)
}
