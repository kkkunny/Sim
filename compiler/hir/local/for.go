package local

import (
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// For 高级循环
type For struct {
	pos      *list.Element[Local]
	cursor   *values.VarDecl
	iterator values.Value
	body     *Block
}

func NewFor(p *Block, cursor *values.VarDecl, iter values.Value) *For {
	return &For{
		cursor:   cursor,
		iterator: iter,
		body:     NewBlock(p),
	}
}

func (self *For) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *For) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *For) Body() *Block {
	return self.body
}

func (self *For) Cursor() *values.VarDecl {
	return self.cursor
}

func (self *For) Iterator() values.Value {
	return self.iterator
}

func (self *For) BlockEndType() BlockEndType {
	switch endType := self.body.BlockEndType(); endType {
	case BlockEndTypeLoopBreak, BlockEndTypeLoopNext:
		return BlockEndTypeNone
	default:
		return endType
	}
}
