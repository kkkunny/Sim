package parse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
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
		return self.parseIfElse()
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

func (self *Parser) parseIfElse() *IfElse {
	begin := self.expectNextIs(token.IF).Position
	cond := self.mustExpr(self.parseOptionExpr(false))
	body := self.parseBlock()
	self.skipSEM()
	var next util.Option[*IfElse]
	if self.skipNextIs(token.ELSE) {
		nextBegin := self.curTok.Position
		self.skipSEM()
		if self.nextIs(token.IF) {
			next = util.Some(self.parseIfElse())
		} else {
			next = util.Some(&IfElse{
				Begin: nextBegin,
				Body:  self.parseBlock(),
			})
		}
	}
	return &IfElse{
		Begin: begin,
		Cond:  util.Some(cond),
		Body:  body,
		Next:  next,
	}
}
