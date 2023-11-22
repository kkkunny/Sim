package parse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"

	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseStmt() ast.Stmt {
	switch self.nextTok.Kind {
	case token.RETURN:
		return self.parseReturn()
	case token.LET:
		return self.parseVariable(nil)
	case token.LBR:
		return self.parseBlock()
	case token.IF:
		return self.parseIfElse()
	case token.LOOP:
		return self.parseLoop()
	case token.BREAK:
		return self.parseBreak()
	case token.CONTINUE:
		return self.parseContinue()
	case token.FOR:
		return self.parseFor()
	default:
		return self.mustExpr(self.parseOptionExpr(true))
	}
}

func (self *Parser) parseBlock() *ast.Block {
	begin := self.expectNextIs(token.LBR).Position
	stmts := linkedlist.NewLinkedList[ast.Stmt]()
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
	return &ast.Block{
		Begin: begin,
		Stmts: stmts,
		End:   end,
	}
}

func (self *Parser) parseReturn() *ast.Return {
	begin := self.expectNextIs(token.RETURN).Position
	value := self.parseOptionExpr(true)
	return &ast.Return{
		Begin: begin,
		Value: value,
	}
}

func (self *Parser) parseIfElse() *ast.IfElse {
	begin := self.expectNextIs(token.IF).Position
	cond := self.mustExpr(self.parseOptionExpr(false))
	body := self.parseBlock()
	var next util.Option[*ast.IfElse]
	if self.skipNextIs(token.ELSE) {
		nextBegin := self.curTok.Position
		if self.nextIs(token.IF) {
			next = util.Some(self.parseIfElse())
		} else {
			next = util.Some(&ast.IfElse{
				Begin: nextBegin,
				Body:  self.parseBlock(),
			})
		}
	}
	return &ast.IfElse{
		Begin: begin,
		Cond:  util.Some(cond),
		Body:  body,
		Next:  next,
	}
}

func (self *Parser) parseLoop() *ast.Loop {
	begin := self.expectNextIs(token.LOOP).Position
	body := self.parseBlock()
	return &ast.Loop{
		Begin: begin,
		Body:  body,
	}
}

func (self *Parser) parseBreak() *ast.Break {
	return &ast.Break{Token: self.expectNextIs(token.BREAK)}
}

func (self *Parser) parseContinue() *ast.Continue {
	return &ast.Continue{Token: self.expectNextIs(token.CONTINUE)}
}

func (self *Parser) parseFor() *ast.For {
	begin := self.expectNextIs(token.FOR).Position
	mut := self.skipNextIs(token.MUT)
	cursor := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.IN)
	iter := self.mustExpr(self.parseOptionExpr(false))
	body := self.parseBlock()
	return &ast.For{
		Begin:     begin,
		CursorMut: mut,
		Cursor:    cursor,
		Iterator:  iter,
		Body:      body,
	}
}