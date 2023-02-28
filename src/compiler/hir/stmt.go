package hir

import "github.com/kkkunny/stl/list"

// Stmt 语句
type Stmt interface {
	stmt()
}

// Block 代码块
type Block struct {
	Stmts *list.SingleLinkedList[Stmt] // 语句
}

func NewBlock(stmts *list.SingleLinkedList[Stmt]) *Block {
	return &Block{
		Stmts: stmts,
	}
}

func (self Block) stmt() {}

// IfElse 条件分支
type IfElse struct {
	Cond  Expr   // 条件
	True  *Block // true分支
	False *Block // false分支（可能为空）
}

func NewIfElse(cond Expr, t, f *Block) *IfElse {
	return &IfElse{
		Cond:  cond,
		True:  t,
		False: f,
	}
}

func (self IfElse) stmt() {}

// Loop 条件循环
type Loop struct {
	Cond Expr   // 条件
	Body *Block // 循环体
}

func NewLoop(cond Expr, body *Block) *Loop {
	return &Loop{
		Cond: cond,
		Body: body,
	}
}

func (self Loop) stmt() {}

// LoopControl 循环控制
type LoopControl interface {
	Stmt
	loopControl()
}

// Break 跳出循环
type Break struct{}

func NewBreak() *Break {
	return &Break{}
}

func (self Break) stmt() {}

func (self Break) loopControl() {}

// Continue 跳过当前循环
type Continue struct{}

func NewContinue() *Continue {
	return &Continue{}
}

func (self Continue) stmt() {}

func (self Continue) loopControl() {}

// Return 函数返回
type Return struct {
	Value Expr // 返回值（可能为空）
}

func NewReturn(v Expr) *Return {
	return &Return{
		Value: v,
	}
}

func (self Return) stmt() {}

// Switch 分支
type Switch struct {
	From       Expr     // 判断值
	CaseValues []Expr   // 分支值
	CaseBodies []*Block // 分支体
	Default    *Block   // 默认分支体（可能为空）
}

func NewSwitch(from Expr, cv []Expr, cb []*Block, d *Block) *Switch {
	return &Switch{
		From:       from,
		CaseValues: cv,
		CaseBodies: cb,
		Default:    d,
	}
}

func (self Switch) stmt() {}

// Variable 变量定义
type Variable struct {
	Typ   Type // 类型
	Value Expr // 值
}

func NewVariable(t Type, v Expr) *Variable {
	return &Variable{
		Typ:   t,
		Value: v,
	}
}

func (self Variable) stmt() {}

func (self Variable) Type() Type {
	return self.Typ
}

func (self Variable) Mutable() bool {
	return true
}

func (self Variable) Immediate() bool {
	return false
}

func (self Variable) ident() {}

// Match 枚举匹配
type Match struct {
	From       Expr     // 判断值
	CaseValues []string // 分支值
	CaseBodies []*Block // 分支体
	Default    *Block   // 默认分支体（可能为空）
}

func NewMatch(from Expr, cv []string, cb []*Block, d *Block) *Match {
	return &Match{
		From:       from,
		CaseValues: cv,
		CaseBodies: cb,
		Default:    d,
	}
}

func (self Match) stmt() {}
