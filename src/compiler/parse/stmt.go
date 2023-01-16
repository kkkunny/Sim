package parse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/list"
)

// Stmt 语句
type Stmt interface {
	Ast
	Stmt()
}

// Block 代码块
type Block struct {
	Pos   utils.Position
	Stmts *list.SingleLinkedList[Stmt]
}

func NewBlock(pos utils.Position) *Block {
	return &Block{
		Pos:   pos,
		Stmts: list.NewSingleLinkedList[Stmt](),
	}
}

func (self Block) Position() utils.Position {
	return self.Pos
}

func (self Block) Stmt() {}

// LoopControl 循环控制
type LoopControl struct {
	Kind lex.Token
}

func NewLoopControl(kind lex.Token) *LoopControl {
	return &LoopControl{Kind: kind}
}

func (self LoopControl) Position() utils.Position {
	return self.Kind.Pos
}

func (self LoopControl) Stmt() {}

// Return 函数返回
type Return struct {
	Pos   utils.Position
	Value Expr // 可能为空
}

func NewReturn(pos utils.Position, v Expr) *Return {
	return &Return{
		Pos:   pos,
		Value: v,
	}
}

func (self Return) Position() utils.Position {
	return self.Pos
}

func (self Return) Stmt() {}

// Variable 变量
type Variable struct {
	Pos   utils.Position
	Type  Type // 可能为空（和value只可能一个为空）
	Name  lex.Token
	Value Expr // 可能为空（和type只可能一个为空）
}

func NewVariable(pos utils.Position, t Type, name lex.Token, v Expr) *Variable {
	return &Variable{
		Pos:   pos,
		Type:  t,
		Name:  name,
		Value: v,
	}
}

func (self Variable) Position() utils.Position {
	return self.Pos
}

func (self Variable) Stmt() {}

// IfElse if else
// cond == nil && next == nil
// cond != nil && next == nil
type IfElse struct {
	Pos  utils.Position
	Cond Expr // 可能为空
	Body *Block
	Next *IfElse // 可能为空
}

func NewIfElse(pos utils.Position, cond Expr, body *Block, next *IfElse) *IfElse {
	return &IfElse{
		Pos:  pos,
		Cond: cond,
		Body: body,
		Next: next,
	}
}

func (self IfElse) Position() utils.Position {
	return self.Pos
}

func (self IfElse) Stmt() {}

// Loop 循环
type Loop struct {
	Pos  utils.Position
	Cond Expr
	Body *Block
}

func NewLoop(pos utils.Position, cond Expr, body *Block) *Loop {
	return &Loop{
		Pos:  pos,
		Cond: cond,
		Body: body,
	}
}

func (self Loop) Position() utils.Position {
	return self.Pos
}

func (self Loop) Stmt() {}

// ****************************************************************

// 语句
func (self *Parser) parseStmt() Stmt {
	switch self.nextTok.Kind {
	case lex.RETURN:
		return self.parseReturn()
	case lex.LET:
		return self.parseVariable()
	case lex.LBR:
		return self.parseBlock()
	case lex.IF:
		return self.parseIfElse()
	case lex.FOR:
		return self.parseFor()
	case lex.BREAK, lex.CONTINUE:
		self.next()
		return NewLoopControl(self.curTok)
	default:
		return self.parseExpr()
	}
}

// 代码块
func (self *Parser) parseBlock() *Block {
	block := NewBlock(utils.Position{})
	begin := self.expectNextIs(lex.LBR).Pos

	for self.skipSem(); !self.nextIs(lex.RBR); self.skipSem() {
		block.Stmts.Add(self.parseStmt())
		self.expectNextIs(lex.SEM)
	}

	end := self.expectNextIs(lex.RBR).Pos
	block.Pos = utils.MixPosition(begin, end)
	return block
}

// 变量
func (self *Parser) parseVariable() *Variable {
	begin := self.expectNextIs(lex.LET).Pos
	name := self.expectNextIs(lex.IDENT)
	var t Type
	var v Expr
	if !self.skipNextIs(lex.COL) {
		self.expectNextIs(lex.ASS)
		v = self.parseExpr()
	} else {
		t = self.parseTypeOrNil()
		if self.skipNextIs(lex.ASS) {
			v = self.parseExpr()
		}
	}
	return NewVariable(utils.MixPosition(begin, self.curTok.Pos), t, name, v)
}

// 函数返回
func (self *Parser) parseReturn() *Return {
	begin := self.expectNextIs(lex.RETURN).Pos
	var value Expr
	if !self.nextIs(lex.SEM) {
		value = self.parseExpr()
	}
	return NewReturn(utils.MixPosition(begin, self.curTok.Pos), value)
}

// ifelse
func (self *Parser) parseIfElse() *IfElse {
	begin := self.expectNextIs(lex.IF).Pos
	cond := self.parseExpr()
	body := self.parseBlock()
	var next *IfElse
	if self.skipNextIs(lex.ELSE) {
		if self.nextIs(lex.IF) {
			next = self.parseIfElse()
		} else {
			nb := self.parseBlock()
			next = NewIfElse(nb.Pos, nil, nb, nil)
		}
	}
	return NewIfElse(utils.MixPosition(begin, self.curTok.Pos), cond, body, next)
}

// 循环
func (self *Parser) parseFor() *Loop {
	begin := self.expectNextIs(lex.FOR).Pos
	cond := self.parseExpr()
	body := self.parseBlock()
	return NewLoop(utils.MixPosition(begin, body.Pos), cond, body)
}
