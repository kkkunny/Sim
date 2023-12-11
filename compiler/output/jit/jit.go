package jit

import "C"
import (
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/runtime"
)

// RunJit jit
func RunJit(module llvm.Module) (uint8, stlerror.Error) {
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
	engine, err := stlerror.ErrorWith(llvm.NewJITCompiler(module, llvm.CodeOptLevelNone))
	if err != nil {
		return 0, err
	}
	mainFn, ok := engine.GetFunction("main")
	if !ok {
		return 0, stlerror.Errorf("can not fond the main function")
	}
	strEqStrFn, ok := engine.GetFunction("sim_runtime_str_eq_str")
	if ok {
		engine.MapGlobalToC(strEqStrFn, runtime.StrEqStr)
	}
	debugFn, ok := engine.GetFunction("sim_runtime_debug")
	if ok {
		engine.MapGlobalToC(debugFn, runtime.Debug)
	}
	checkNull, ok := engine.GetFunction("sim_runtime_check_null")
	if ok {
		engine.MapGlobalToC(checkNull, runtime.CheckNull)
	}
	return uint8(engine.RunFunction(mainFn).Integer(false)), nil
}
