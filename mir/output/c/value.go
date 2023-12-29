package c

import (
	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) codegenValue(ir mir.Value)string{
	switch value := ir.(type) {
	case mir.Const:
		return self.codegenConst(value)
	default:
		return self.values.Get(ir)
	}
}
