package interpret

import (
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/execution"
	"github.com/kkkunny/Sim/runtime"
)

// Interpret 解释执行
func Interpret(module *mir.Module)(uint8, stlerror.Error){
	engine, err := execution.NewExecutionEngine(module)
	if err != nil{
		return 1, err
	}
	for _, runtimeFn := range runtime.RuntimeFuncList{
		err = engine.MapFunctionIgnoreNotFind(runtimeFn.Name, runtimeFn.To)
		if err != nil{
			return 1, err
		}
	}
	return engine.RunMain()
}
