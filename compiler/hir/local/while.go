package local

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

// While 循环
type While struct {
	cond hir.Value
	body *Block
}

func NewWhile(cond hir.Value, body *Block) *While {
	return &While{
		cond: cond,
		body: body,
	}
}

func (self *While) Local() {
	return
}

func (self *While) Cond() hir.Value {
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

func (self *While) loop() {
	return
}
