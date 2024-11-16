package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// IfElse 条件语句
type IfElse struct {
	cond hir.Value
	body *Block
	next *IfElse
}

func NewIfElse(body *Block, cond ...hir.Value) *IfElse {
	return &IfElse{
		cond: stlslices.Last(cond),
		body: body,
	}
}

func (self *IfElse) Local() {
	return
}

func (self *IfElse) Cond() (hir.Value, bool) {
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
