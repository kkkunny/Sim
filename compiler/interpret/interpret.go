package interpret

import (
	"reflect"
	"strings"

	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlerror "github.com/kkkunny/stl/error"
)

// Engine 执行器
type Engine struct {
	module llvm.Module
	target *llvm.Target
	jiter  *llvm.ExecutionEngine
}

func NewExecutionEngine(module llvm.Module) (*Engine, stlerror.Error) {
	engine, err := stlerror.ErrorWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	if err != nil {
		return nil, err
	}
	stlerror.Must(llvm.InitializeNativeAsmParser())
	stlerror.Must(llvm.InitializeNativeAsmPrinter())
	return &Engine{
		module: module,
		target: stlbasic.IgnoreWith(module.GetTarget()),
		jiter:  engine,
	}, nil
}

// MapFunction 映射函数 interpreter
func (self *Engine) MapFunction(name string, to any) stlerror.Error {
	toVal := reflect.ValueOf(to)
	toFt := toVal.Type()
	if toFt.Kind() != reflect.Func {
		return stlerror.Errorf("expect a function")
	}

	f, ok := self.module.GetFunction(name)
	if !ok {
		return stlerror.Errorf("unknown function which named `%s`", name)
	}
	ft := f.FunctionType()
	switch {
	case self.target.IsWindows():
		if stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.ArrayType](ft.ReturnType()) || stlslices.Any(ft.Params(), func(i int, p llvm.Type) bool {
			return stlbasic.Is[llvm.StructType](p) || stlbasic.Is[llvm.ArrayType](p)
		}) {
			retTypes, paramTypes := make([]reflect.Type, 0, 1), make([]reflect.Type, 0, toFt.NumIn())
			if stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.ArrayType](ft.ReturnType()) {
				paramTypes = append(paramTypes, reflect.PtrTo(toFt.Out(0)))
			} else if toFt.NumOut() != 0 {
				retTypes = append(retTypes, toFt.Out(0))
			}
			for i, pt := range ft.Params() {
				if stlbasic.Is[llvm.StructType](pt) || stlbasic.Is[llvm.ArrayType](pt) {
					paramTypes = append(paramTypes, reflect.PtrTo(toFt.In(i)))
				} else {
					paramTypes = append(paramTypes, toFt.In(i))
				}
			}
			toFt = reflect.FuncOf(paramTypes, retTypes, false)

			srcToVal := toVal
			toVal = reflect.MakeFunc(toFt, func(args []reflect.Value) (results []reflect.Value) {
				callArgs := make([]reflect.Value, 0, len(args))
				skip := stlbasic.Ternary(stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.ArrayType](ft.ReturnType()), 1, 0)
				for i, pt := range ft.Params() {
					if stlbasic.Is[llvm.StructType](pt) || stlbasic.Is[llvm.ArrayType](pt) {
						callArgs = append(callArgs, args[skip+i].Elem())
					} else {
						callArgs = append(callArgs, args[skip+i])
					}
				}

				rets := srcToVal.Call(callArgs)

				if stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.ArrayType](ft.ReturnType()) {
					args[0].Elem().Set(rets[0])
					return nil
				} else if toFt.NumOut() != 0 {
					return rets
				} else {
					return nil
				}
			})
			to = toVal.Interface()
		}
		fallthrough
	default:
		return stlerror.ErrorWrap(self.jiter.MapFunctionToGo(name, to))
	}
}

func (self *Engine) MapFunctionIgnoreNotFind(name string, to any) stlerror.Error {
	err := self.MapFunction(name, to)
	if err != nil && !strings.Contains(err.Error(), "unknown function") {
		return err
	}
	return nil
}

func (self *Engine) RunMain() (uint8, stlerror.Error) {
	mainFn, ok := self.module.GetFunction("main")
	if !ok {
		return 1, stlerror.Errorf("can not find the main function")
	}
	return self.jiter.RunMainFunction(mainFn, nil, nil), nil
}
