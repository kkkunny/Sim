package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// IfElse 条件语句
type IfElse struct {
	cond values.Value
	body *Block
	next *IfElse
}

func NewIfElse(body *Block, cond ...values.Value) *IfElse {
	return &IfElse{
		cond: stlslices.Last(cond),
		body: body,
	}
}

func (self *IfElse) local() {
	return
}

func (self *IfElse) Cond() (values.Value, bool) {
	return self.cond, self.cond != nil
}

func (self *IfElse) Body() *Block {
	return self.body
}

func (self *IfElse) SetNext(ifelse *IfElse) *IfElse {
	self.next = ifelse
	return ifelse
}

func (self *IfElse) Next() (*IfElse, bool) {
	return self.next, self.next != nil
}

func (self *IfElse) BlockEndType() BlockEndType {
	endType := self.body.BlockEndType()
	if self.next == nil {
		return BlockEndTypeNone
	}
	return min(endType, self.next.BlockEndType())
}
