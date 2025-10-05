package interpret

import (
	"errors"

	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/runtime"
)

// Interpret 解释执行
func Interpret(module llvm.Module) (uint8, error) {
	if _, ok := module.GetFunction("interpret_main"); ok {
		return 1, errors.New("repeated interpret_main function")
	}

	// 插入main函数
	ctx := module.Context()
	builder := ctx.NewBuilder()
	interpretMainFn := module.NewFunction("interpret_main", ctx.FunctionType(false, ctx.IntegerType(8)))
	entryBlock := interpretMainFn.NewBlock("")
	builder.MoveToAfter(entryBlock)

	// 调用init函数
	if initFn, ok := module.GetFunction("sim_runtime_init"); ok {
		builder.CreateCall("", initFn.FunctionType(), initFn)
	}

	// 调用main函数
	if mainFn, ok := module.GetFunction("main"); ok {
		ret := builder.CreateCall("", mainFn.FunctionType(), mainFn)
		builder.CreateRet(ret)
	} else {
		builder.CreateRet(nil)
	}

	// 新建运行器，并绑定外部变量、函数
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

	// 运行程序
	return engine.RunMain()
}
