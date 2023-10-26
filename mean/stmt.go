package mean

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/util"
)

// Stmt 语句
type Stmt interface {
	stmt()
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
