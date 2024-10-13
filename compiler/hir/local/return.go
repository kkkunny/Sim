package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// Return 函数返回
type Return struct {
	pos   *list.Element[Local]
	value values.Value
}

func NewReturn(v ...values.Value) *Return {
	return &Return{
		value: stlslices.Last(v),
	}
}

func (self *Return) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Return) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Return) Value() values.Value {
	return self.value
}

func (self *Return) BlockEndType() BlockEndType {
	return BlockEndTypeFuncRet
}
