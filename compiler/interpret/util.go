package interpret

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/runtime"
)

// Interpret 解释执行
func Interpret(module llvm.Module) (uint8, error) {
	engine, err := NewExecutionEngine(module)
	if err != nil {
		return 1, err
	}
	for _, runtimeFn := range runtime.FuncList {
		err = engine.MapFunctionIgnoreNotFind(runtimeFn.Name, runtimeFn.C, runtimeFn.To)
		if err != nil {
			return 1, err
		}
	}
	return engine.RunMain()
}
