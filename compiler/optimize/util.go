package optimize

import (
	"github.com/kkkunny/go-llvm"
)

// Opt 优化
func Opt(module llvm.Module) (llvm.Module, error) {
	module.AutoOpt(llvm.OptLevelO2)
	return module, nil
}
