package ast

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
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
	Token token.Token
}

func (self *Return) Position() reader.Position {
	return self.Token.Position
}

func (self *Return) stmt() {}
