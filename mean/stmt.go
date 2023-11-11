package mean

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/util"
)

// Stmt 语句
type Stmt interface {
	stmt()
}

// JumpOut 跳出
type JumpOut uint8

// 值越大优先级越大
const (
	JumpOutNone JumpOut = iota
	JumpOutLoop
	JumpOutReturn
)

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

// IfElse if else
type IfElse struct {
	Cond util.Option[Expr]
	Body *Block
	Next util.Option[*IfElse]
}

func (self *IfElse) IsIf() bool {
	return self.Cond.IsSome()
}

func (self *IfElse) IsElse() bool {
	return self.Cond.IsNone()
}

func (self *IfElse) HasElse() bool {
	if self.IsElse() {
		return true
	}
	if nextIfElse, ok := self.Next.Value(); ok {
		return nextIfElse.HasElse()
	}
	return false
}

func (*IfElse) stmt() {}

type Loop struct {
	Body *Block
}

func (*Loop) stmt() {}

type Break struct {
	Loop *Loop
}

func (*Break) stmt() {}
