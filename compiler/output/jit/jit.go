package jit

import (
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
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
	return uint8(engine.RunFunction(*mainFn).Integer(false)), nil
}