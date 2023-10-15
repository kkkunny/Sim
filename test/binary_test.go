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

func TestAdd(t *testing.T) {
	code := `
func main()isize{
    return 1+1-2
}
`
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}

func TestSub(t *testing.T) {
	code := `
func main()isize{
    return 2-2
}
`
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}

func TestMul(t *testing.T) {
	code := `
func main()isize{
    return 10*0
}
`
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}

func TestDiv(t *testing.T) {
	code := `
func main()isize{
    return 0/10
}
`
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}

func TestRem(t *testing.T) {
	code := `
func main()isize{
    return 10%2
}
`
	r := stlerror.MustWith(reader.NewReaderFromString("test.sim", code))
	module := codegen.New(analyse.New(parse.New(lex.New(r)))).Codegen()
	engine := stlerror.MustWith(llvm.NewInterpreter(module))
	ret := engine.RunFunction(engine.FindFunction("main"), nil).Int(true)
	stltest.AssertEq(t, ret, 0)
}
