package hir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
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
	Value optional.Optional[Expr]
}

func NewReturn(f CallableDef, value ...Expr) *Return {
	return &Return{
		Func: f,
		Value: stlbasic.TernaryAction(stlslices.Empty(value), func() optional.Optional[Expr] {
			return optional.None[Expr]()
		}, func() optional.Optional[Expr] {
			return optional.Some[Expr](NewMoveOrCopy(stlslices.Last(value)))
		}),
	}
}

func (*Return) stmt() {}

func (*Return) out() {}

// IfElse if else
type IfElse struct {
	Cond optional.Optional[Expr]
	Body *Block
	Next optional.Optional[*IfElse]
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
	Other optional.Optional[*Block]
}

func (*Match) stmt() {}

// LocalVarDef 局部变量定义
type LocalVarDef struct {
	VarDecl
	Value   optional.Optional[Expr]
	Escaped bool
}

func NewLocalVarDef(mut bool, name string, t Type, value ...Expr) *LocalVarDef {
	return &LocalVarDef{
		VarDecl: VarDecl{
			Mut:  mut,
			Name: name,
			Type: t,
		},
		Value: stlbasic.TernaryAction(stlslices.Empty(value), func() optional.Optional[Expr] {
			return optional.None[Expr]()
		}, func() optional.Optional[Expr] {
			return optional.Some[Expr](NewMoveOrCopy(stlslices.Last(value)))
		}),
	}
}

func (*LocalVarDef) stmt() {}

func (self *LocalVarDef) SetValue(value Expr) {
	self.Value = optional.Some[Expr](NewMoveOrCopy(value))
}

// MultiLocalVarDef 多局部变量定义
type MultiLocalVarDef struct {
	Vars  []*LocalVarDef
	Value Expr
}

func (*MultiLocalVarDef) stmt() {}
