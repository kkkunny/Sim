package llvm

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/mir"
)

func (self *LLVMOutputer) codegenValue(ir mir.Value)llvm.Value{
	switch value := ir.(type) {
	case mir.Const:
		return self.codegenConst(value)
	default:
		return self.values.Get(ir)
	}
}
