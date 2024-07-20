package interpret

import (
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/runtime/extern"
)

// Interpret 解释执行
func Interpret(module llvm.Module) (uint8, stlerror.Error) {
	engine, err := NewExecutionEngine(module)
	if err != nil {
		return 1, err
	}
	for _, runtimeFn := range extern.FuncList {
		err = engine.MapFunctionIgnoreNotFind(runtimeFn.Name, runtimeFn.To)
		if err != nil {
			return 1, err
		}
	}
	return engine.RunMain()
}
