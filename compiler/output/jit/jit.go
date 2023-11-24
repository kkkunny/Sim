package jit

import "C"
import (
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/runtime"
)

// RunJit jit
func RunJit(module llvm.Module) (uint8, stlerror.Error) {
	engine, err := stlerror.ErrorWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	if err != nil {
		return 0, err
	}
	mainFn := module.GetFunction("main")
	if mainFn == nil {
		return 0, stlerror.Errorf("can not fond the main function")
	}
	strEqStrFn := module.GetFunction("sim_runtime_str_eq_str")
	if strEqStrFn != nil {
		engine.MapGlobal(strEqStrFn, runtime.StrEqStr)
	}
	debugFn := module.GetFunction("sim_runtime_debug")
	if debugFn != nil {
		engine.MapGlobal(debugFn, runtime.Debug)
	}
	return uint8(engine.RunFunction(*mainFn).Integer(false)), nil
}
