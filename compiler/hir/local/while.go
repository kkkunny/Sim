package local

import (
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// While 循环
type While struct {
	pos  *list.Element[Local]
	cond values.Value
	body *Block
}

func NewWhile(p *Block, c values.Value) *While {
	return &While{
		cond: c,
		body: NewBlock(p),
	}
}

func (self *While) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *While) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
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
