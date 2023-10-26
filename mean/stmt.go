package mean

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/util"
)

// Stmt 语句
type Stmt interface {
	stmt()
}

// SkipOut 跳出
type SkipOut interface {
	Stmt
	out()
}

// Block 代码块
type Block struct {
	Stmts linkedlist.LinkedList[Stmt]
}

func (*Block) stmt() {}

// Return 函数返回
type Return struct {
	Value util.Option[Expr]
}

func (*Return) stmt() {}

func (*Return) out() {}

// Variable 变量定义
type Variable struct {
	Type  Type
	Name  string
	Value Expr
}

func (*Variable) stmt() {}

func (self *Variable) GetType() Type {
	return self.Type
}

func (*Variable) ident() {}

// If if
type If struct {
	Cond Expr
	Body *Block
}

func (*If) stmt() {}
