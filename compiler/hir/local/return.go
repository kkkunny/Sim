package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// Return 函数返回
type Return struct {
	value hir.Value
}

func NewReturn(v ...hir.Value) *Return {
	return &Return{
		value: stlslices.Last(v),
	}
}

func (self *Return) Local() {
	return
}

func (self *Return) Value() (hir.Value, bool) {
	return self.value, self.value != nil
}

func (self *Return) BlockEndType() BlockEndType {
	return BlockEndTypeFuncRet
}
