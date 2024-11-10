package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// Return 函数返回
type Return struct {
	value values.Value
}

func NewReturn(v ...values.Value) *Return {
	return &Return{
		value: stlslices.Last(v),
	}
}

func (self *Return) local() {
	return
}

func (self *Return) Value() (values.Value, bool) {
	return self.value, self.value != nil
}

func (self *Return) BlockEndType() BlockEndType {
	return BlockEndTypeFuncRet
}
