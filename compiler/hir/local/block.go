package local

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/list"
	stlval "github.com/kkkunny/stl/value"
)

type Block struct {
	parent either.Either[CallableDef, *Block]
	stmts  *list.List[Local]
	idents hashmap.HashMap[string, any]

	inLoop bool
}

func NewFuncBody(f CallableDef) *Block {
	return &Block{
		parent: either.Left[CallableDef, *Block](f),
		stmts:  list.New[Local](),
		idents: hashmap.StdWith[string, any](),
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
	for cursor := self.stmts.Front(); cursor != nil; cursor = cursor.Next() {
		if end, ok := cursor.Value.(blockEnd); ok {
			if endType := end.BlockEndType(); endType > BlockEndTypeNone {
				return endType
			}
		}
	}
	return BlockEndTypeNone
}

func (self *Block) local() {
	return
}

func (self *Block) Append(block *Block) {
	self.stmts.PushBackList(block.stmts)
	for iter := self.idents.Iterator(); iter.Next(); {
		pair := iter.Value()
		self.idents.Set(pair.E1(), pair.E2())
	}
}

func (self *Block) PushBack(l Local) *Block {
	self.stmts.PushBack(l)
	return self
}

func (self *Block) SetIdent(name string, ident any) bool {
	self.idents.Set(name, ident)
	return true
}

func (self *Block) GetIdent(name string, allowLinkedPkgs ...bool) (any, bool) {
	v := self.idents.Get(name)
	if v != nil {
		return v, true
	}
	parent := stlval.TernaryAction(self.parent.IsLeft(), func() Scope {
		return stlval.IgnoreWith(self.parent.Left()).Parent()
	}, func() Scope {
		return stlval.IgnoreWith(self.parent.Right())
	})
	return parent.GetIdent(name, allowLinkedPkgs...)
}

func (self *Block) Belong() CallableDef {
	f, ok := self.parent.Left()
	if ok {
		return f
	}
	return stlval.IgnoreWith(self.parent.Right()).Belong()
}

func (self *Block) SetInLoop(v bool) {
	self.inLoop = v
}

func (self *Block) InLoop() bool {
	if self.inLoop {
		return true
	}
	parent, ok := self.parent.Right()
	if !ok {
		return false
	}
	return parent.InLoop()
}
