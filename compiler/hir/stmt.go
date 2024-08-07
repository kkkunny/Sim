package hir

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/compiler/util"
)

// Stmt 语句
type Stmt interface {
	stmt()
}

// BlockEof 代码块结束符
type BlockEof uint8

// 值越大优先级越大
const (
	BlockEofNone BlockEof = iota
	BlockEofNextLoop
	BlockEofBreakLoop
	BlockEofReturn
)

// Block 代码块
type Block struct {
	Stmts linkedlist.LinkedList[Stmt]
}

func (*Block) stmt() {}

// Return 函数返回
type Return struct {
	Func  CallableDef
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

type Loop interface {
	Stmt
	loop()
}

type While struct {
	Cond Expr
	Body *Block
}

func (*While) stmt() {}
func (*While) loop() {}

type Break struct {
	Loop Loop
}

func (*Break) stmt() {}

type Continue struct {
	Loop Loop
}

func (*Continue) stmt() {}

type For struct {
	Cursor   *LocalVarDef
	Iterator Expr
	Body     *Block
}

func (*For) stmt() {}
func (*For) loop() {}

type MatchCase struct {
	Name  string
	Elems []*Param
	Body  *Block
}

type Match struct {
	Value Expr
	Cases linkedhashmap.LinkedHashMap[string, *MatchCase]
	Other util.Option[*Block]
}

func (*Match) stmt() {}

// LocalVarDef 局部变量定义
type LocalVarDef struct {
	VarDecl
	Value   Expr
	Escaped bool
}

func (*LocalVarDef) stmt() {}

// MultiLocalVarDef 多局部变量定义
type MultiLocalVarDef struct {
	Vars  []*LocalVarDef
	Value Expr
}

func (*MultiLocalVarDef) stmt() {}
