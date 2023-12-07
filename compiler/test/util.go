package test

import (
	"path/filepath"
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/config"
	_ "github.com/kkkunny/Sim/config"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/output/jit"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

func init() {
	config.ROOT = filepath.Dir(config.ROOT)
	stlerror.Must(llvm.InitializeNativeTarget())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
}

func assertRetEq(t *testing.T, code string, expect uint8, skips ...uint) {
	var skip uint
	if len(skips) > 0{
		skip = skips[0]
	}
	path := stlerror.MustWith(filepath.Abs(stlerror.MustWith(util.GetFileName(skip+1))))
	target := stlerror.MustWith(llvm.NativeTarget())
	r := stlerror.MustWith(reader.NewReaderFromString(path, code))
	module := codegen.New(target, analyse.New(parse.New(lex.New(r)).Parse(), target)).Codegen()
	stlerror.Must(module.Verify())
	stltest.AssertEq(t, stlerror.MustWith(jit.RunJit(module)), expect)
}

func assertRetEqZero(t *testing.T, code string) {
	assertRetEq(t, code, 0, 1)
}
