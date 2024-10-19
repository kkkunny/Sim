package local

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/list"
	stlval "github.com/kkkunny/stl/value"
)

type Block struct {
	pos    *list.Element[Local]
	parent either.Either[CallableDef, *Block]
	stmts  *list.List[Local]
}

func NewFuncBody(f CallableDef) *Block {
	return &Block{
		parent: either.Left[CallableDef, *Block](f),
		stmts:  list.New[Local](),
	}
}

func NewBlock(p *Block) *Block {
	return &Block{
		parent: either.Right[CallableDef, *Block](p),
		stmts:  list.New[Local](),
	}
}

func (self *Block) Parent() either.Either[CallableDef, *Block] {
	return self.parent
}

func (self *Block) CallableDef() CallableDef {
	parentFunc, ok := self.parent.Left()
	if ok {
		return parentFunc
	}
	return stlval.IgnoreWith(self.parent.Right()).CallableDef()
}

func (self *Block) Stmts() *list.List[Local] {
	return self.stmts
}

func (self *Block) Empty() bool {
	return self.stmts.Len() == 0
}

func (self *Block) HasEnd() bool {
	return self.BlockEndType() != BlockEndTypeNone
}

func (self *Block) BlockEndType() BlockEndType {
	// TODO 完善
	for cursor := self.stmts.Front(); cursor != nil; cursor = cursor.Next() {
		if blockEnd, ok := cursor.Value.(blockEnd); ok {
			return blockEnd.BlockEndType()
		}
	}
	return BlockEndTypeNone
}

func (self *Block) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Block) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Block) Append(l Local) *Block {
	l.setPosition(self.stmts.PushBack(l))
	return self
}

func (self *Block) Remove(l Local) *Block {
	pos, ok := l.position()
	if !ok {
		return self
	}
	self.stmts.Remove(pos)
	return self
}
