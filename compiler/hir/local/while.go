package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// While 循环
type While struct {
	cond values.Value
	body *Block
}

func NewWhile(cond values.Value, body *Block) *While {
	return &While{
		cond: cond,
		body: body,
	}
}

func (self *While) local() {
	return
}

func (self *While) Cond() values.Value {
	return self.cond
}

func (self *While) Body() *Block {
	return self.body
}

func (self *While) BlockEndType() BlockEndType {
	switch endType := self.body.BlockEndType(); endType {
	case BlockEndTypeLoopBreak, BlockEndTypeLoopNext:
		return BlockEndTypeNone
	default:
		return endType
	}
}
