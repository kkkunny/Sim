package test

import (
	"testing"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/interpret"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/mir"
	modulePass "github.com/kkkunny/Sim/mir/pass/module"
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
	module := codegen_ir.New(mir.DefaultTarget(), analyse.New(parse.New(lex.New(r)).Parse()).Analyse()).Codegen()
	modulePass.Run(module, modulePass.DeadCodeElimination)
	ret := stlerror.MustWith(interpret.Interpret(module))
	stltest.AssertEq(t, ret, expect)
}

func assertRetEqZero(t *testing.T, code string) {
	assertRetEq(t, code, 0, 1)
}
