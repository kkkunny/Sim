package test

import (
	"testing"

	"github.com/kkkunny/llvm"
	stlerror "github.com/kkkunny/stl/error"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

var code = `
func main()isize{
    return 0
}
`

func TestReturn(t *testing.T) {
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}
