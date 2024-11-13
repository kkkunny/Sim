package interpret

import (
	"reflect"
	"strings"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
)

// Engine 执行器
type Engine struct {
	module llvm.Module
	target llvm.Target
	jiter  *llvm.ExecutionEngine
}

func NewExecutionEngine(module llvm.Module) (*Engine, error) {
	target, err := stlerror.ErrorWith(module.GetTarget())
	if err != nil {
		return nil, err
	}
	engine, err := stlerror.ErrorWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	if err != nil {
		return nil, err
	}
	stlerror.Must(llvm.InitializeNativeAsmParser())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	return &Engine{
		module: module,
		target: target,
		jiter:  engine,
	}, nil
}

// MapFunction 映射函数 interpreter
func (self *Engine) MapFunction(name string, to any) error {
	toVal := reflect.ValueOf(to)
	toFt := toVal.Type()
	if toFt.Kind() != reflect.Func {
		return stlerror.Errorf("expect a function")
	}

	return stlerror.ErrorWrap(self.jiter.MapFunctionToGo(name, to))
}

func (self *Engine) MapFunctionIgnoreNotFind(name string, to any) error {
	err := self.MapFunction(name, to)
	if err != nil && !strings.Contains(err.Error(), "unknown function") {
		return err
	}
	return nil
}

func (self *Engine) RunMain() (uint8, error) {
	initFn, ok := self.module.GetFunction("sim_runtime_init")
	if ok {
		_ = self.jiter.RunFunction(initFn)
	}
	mainFn, ok := self.module.GetFunction("main")
	if !ok {
		return 1, stlerror.Errorf("can not find the main function")
	}
	return self.jiter.RunMainFunction(mainFn, nil, nil), nil
}
