package ast

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

// Stmt 语句ast
type Stmt interface {
	Ast
	stmt()
}

// Block 代码块
type Block struct {
	Begin reader.Position
	Stmts linkedlist.LinkedList[Stmt]
	End   reader.Position
}

func (self *Block) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Block) stmt() {}

type Return struct {
	Begin reader.Position
	Value util.Option[Expr]
}

func (self *Return) Position() reader.Position {
	v, ok := self.Value.Value()
	if !ok {
		return self.Begin
	}
	return reader.MixPosition(self.Begin, v.Position())
}

func (self *Return) stmt() {}
