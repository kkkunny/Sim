package hir

import "github.com/kkkunny/stl/types"

// Block 代码块
type Block struct {
	Stmts []Stmt
}

func (self Block) stmt() {}

// Stmt 语句
type Stmt interface {
	stmt()
}

// Return 函数返回
type Return struct {
	Value Expr
}

func (self Return) stmt() {}

// Variable 变量
type Variable struct {
	Type  Type
	Value Expr
}

func (self Variable) stmt() {}

func (self Variable) ident() {}

func (self Variable) GetType() Type {
	return self.Type
}

func (self Variable) GetMut() bool {
	return true
}

func (self Variable) IsTemporary() bool {
	return false
}

// IfElse 条件分支
type IfElse struct {
	Cond        Expr
	True, False *Block
}

func (self IfElse) stmt() {}

// Loop 循环
type Loop struct {
	Cond Expr
	Body *Block
}

func (self Loop) stmt() {}

// LoopControl 循环
type LoopControl struct {
	Type string
}

func (self LoopControl) stmt() {}

// Switch 分支
type Switch struct {
	From    Expr
	Cases   []types.Pair[Expr, *Block]
	Default *Block // 可能为空
}

func (self Switch) stmt() {}
