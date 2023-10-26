package parse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseStmt() Stmt {
	switch self.nextTok.Kind {
	case token.RETURN:
		return self.parseReturn()
	case token.LET:
		return self.parseVariable()
	case token.LBR:
		return self.parseBlock()
	case token.IF:
		return self.parseIf()
	default:
		return self.mustExpr(self.parseOptionExpr(true))
	}
}

func (self *Parser) parseBlock() *Block {
	begin := self.expectNextIs(token.LBR).Position
	stmts := linkedlist.NewLinkedList[Stmt]()
	for {
		self.skipSEM()
		if self.nextIs(token.RBR) {
			break
		}

		stmts.PushBack(self.parseStmt())

		if !self.nextIs(token.RBR) {
			self.expectNextIs(token.SEM)
		}
	}
	end := self.expectNextIs(token.RBR).Position
	return &Block{
		Begin: begin,
		Stmts: stmts,
		End:   end,
	}
}

func (self *Parser) parseReturn() *Return {
	begin := self.expectNextIs(token.RETURN).Position
	value := self.parseOptionExpr(true)
	return &Return{
		Begin: begin,
		Value: value,
	}
}

func (self *Parser) parseVariable() *Variable {
	begin := self.expectNextIs(token.LET).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.COL)
	typ := self.parseType()
	self.expectNextIs(token.ASS)
	value := self.mustExpr(self.parseOptionExpr(true))
	return &Variable{
		Begin: begin,
		Name:  name,
		Type:  typ,
		Value: value,
	}
}

func (self *Parser) parseIf() *If {
	begin := self.expectNextIs(token.IF).Position
	cond := self.mustExpr(self.parseOptionExpr(false))
	body := self.parseBlock()
	return &If{
		Begin: begin,
		Cond:  cond,
		Body:  body,
	}
}
