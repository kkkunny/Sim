package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// For 高级循环
type For struct {
	cursor *values.VarDecl
	iter   values.Value
	body   *Block
}

func NewFor(cursor *values.VarDecl, iter values.Value, body *Block) *For {
	return &For{
		cursor: cursor,
		iter:   iter,
		body:   body,
	}
}

func (self *For) local() {
	return
}

func (self *For) Body() *Block {
	return self.body
}

func (self *For) Cursor() *values.VarDecl {
	return self.cursor
}

func (self *For) Iterator() values.Value {
	return self.iter
}

func (self *For) BlockEndType() BlockEndType {
	switch endType := self.body.BlockEndType(); endType {
	case BlockEndTypeLoopBreak, BlockEndTypeLoopNext:
		return BlockEndTypeNone
	default:
		return endType
	}
}
