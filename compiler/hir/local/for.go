package local

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// Loop 循环
type Loop interface {
	hir.Local
	loop()
}

// For 高级循环
type For struct {
	cursor values.VarDecl
	iter   hir.Value
	body   *Block
}

func NewFor(cursor values.VarDecl, iter hir.Value, body *Block) *For {
	return &For{
		cursor: cursor,
		iter:   iter,
		body:   body,
	}
}

func (self *For) Local() {
	return
}

func (self *For) Body() *Block {
	return self.body
}

func (self *For) Cursor() values.VarDecl {
	return self.cursor
}

func (self *For) Iter() hir.Value {
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

func (self *For) loop() {
	return
}
