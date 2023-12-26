package execution

import (
	"reflect"
	"strings"

	llvm2 "github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	stlerror "github.com/kkkunny/stl/error"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
)

// Engine 执行器
type Engine struct {
	module *mir.Module
	llvmModule llvm2.Module
	target *llvm2.Target
	jiter *llvm2.ExecutionEngine
}

func NewExecutionEngine(module *mir.Module)(*Engine, stlerror.Error){
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(module)
	engine, err := stlerror.ErrorWith(llvm2.NewJITCompiler(outputer.Module(), llvm2.CodeOptLevelNone))
	if err != nil{
		return nil, err
	}
	initTarget(module.Context().Target())
	return &Engine{
		module: module,
		llvmModule: outputer.Module(),
		target: outputer.Target(),
		jiter: engine,
	}, nil
}

// MapFunction 映射函数 interpreter
func (self *Engine) MapFunction(name string, to any)stlerror.Error{
	toVal := reflect.ValueOf(to)
	toFt := toVal.Type()
	if toFt.Kind() != reflect.Func{
		return stlerror.Errorf("expect a function")
	}

	fir, ok := self.module.NamedFunction(name)
	if !ok{
		return stlerror.Errorf("unknown function which named `%s`", name)
	}
	ftir := fir.Type().(mir.FuncType)
	switch {
	case self.target.IsWindows():
		if stlbasic.Is[mir.StructType](ftir.Ret()) || stlbasic.Is[mir.ArrayType](ftir.Ret()) || stlslices.Any(ftir.Params(), func(i int, et mir.Type) bool {
			return stlbasic.Is[mir.StructType](et) || stlbasic.Is[mir.ArrayType](et)
		}){
			retTypes, paramTypes := make([]reflect.Type, 0, 1), make([]reflect.Type, 0, toFt.NumIn())
			if stlbasic.Is[mir.StructType](ftir.Ret()) || stlbasic.Is[mir.ArrayType](ftir.Ret()){
				paramTypes = append(paramTypes, reflect.PtrTo(toFt.Out(0)))
			}else if toFt.NumOut() != 0{
				retTypes = append(retTypes, toFt.Out(0))
			}
			for i, pt := range ftir.Params(){
				if stlbasic.Is[mir.StructType](pt) || stlbasic.Is[mir.ArrayType](pt){
					paramTypes = append(paramTypes, reflect.PtrTo(toFt.In(i)))
				}else{
					paramTypes = append(paramTypes, toFt.In(i))
				}
			}
			toFt = reflect.FuncOf(paramTypes, retTypes, false)

			srcToVal := toVal
			toVal = reflect.MakeFunc(toFt, func(args []reflect.Value) (results []reflect.Value) {
				callArgs := make([]reflect.Value, 0, len(args))
				skip := stlbasic.Ternary(stlbasic.Is[mir.StructType](ftir.Ret()) || stlbasic.Is[mir.ArrayType](ftir.Ret()), 1, 0)
				for i, pt := range ftir.Params(){
					if stlbasic.Is[mir.StructType](pt) || stlbasic.Is[mir.ArrayType](pt){
						callArgs = append(callArgs, args[skip+i].Elem())
					}else{
						callArgs = append(callArgs, args[skip+i])
					}
				}

				rets := srcToVal.Call(callArgs)

				if stlbasic.Is[mir.StructType](ftir.Ret()) || stlbasic.Is[mir.ArrayType](ftir.Ret()){
					args[0].Elem().Set(rets[0])
					return nil
				}else if toFt.NumOut() != 0{
					return rets
				}else{
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

func (self *Engine) MapFunctionIgnoreNotFind(name string, to any)stlerror.Error{
	err := self.MapFunction(name, to)
	if err != nil && !strings.Contains(err.Error(), "unknown function"){
		return err
	}
	return nil
}

func (self *Engine) RunMain()(uint8, stlerror.Error){
	mainFn, ok := self.llvmModule.GetFunction("main")
	if !ok{
		return 1, stlerror.Errorf("can not find the main function")
	}
	return self.jiter.RunMainFunction(mainFn, nil, nil), nil
}
