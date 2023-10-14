package ast

import "github.com/kkkunny/Sim/reader"

// Stmt 语句ast
type Stmt interface {
	Ast
	stmt()
}

// Block 代码块
type Block struct {
	Begin reader.Position
	End   reader.Position
}

func (self Block) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self Block) stmt() {}
