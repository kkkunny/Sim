package parse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/token"
	"github.com/kkkunny/Sim/compiler/util"
)

func (self *Parser) parseStmt() ast.Stmt {
	switch self.nextTok.Kind {
	case token.RETURN:
		return self.parseReturn()
	case token.LET:
		return self.parseVariable(false, nil, nil)
	case token.LBR:
		return self.parseBlock()
	case token.IF:
		return self.parseIfElse()
	case token.BREAK:
		return self.parseBreak()
	case token.CONTINUE:
		return self.parseContinue()
	case token.FOR:
		return self.parseFor()
	case token.MATCH:
		return self.parseMatch()
	case token.WHILE:
		return self.parseWhile()
	default:
		return self.mustExpr(self.parseOptionExpr(true))
	}
}

func (self *Parser) parseStmtList(end ...token.Kind) (stmts linkedlist.LinkedList[ast.Stmt]) {
	for {
		self.skipSEM()
		if self.nextIn(end...) {
			break
		}

		stmts.PushBack(self.parseStmt())

		if !self.nextIn(end...) {
			self.expectNextIs(token.SEM)
		}
	}
	return stmts
}

func (self *Parser) parseBlock() *ast.Block {
	begin := self.expectNextIs(token.LBR).Position
	stmts := self.parseStmtList(token.RBR)
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

func (self *Parser) parseWhile() *ast.While {
	begin := self.expectNextIs(token.WHILE).Position
	cond := self.mustExpr(self.parseOptionExpr(false))
	body := self.parseBlock()
	return &ast.While{
		Begin: begin,
		Cond:  cond,
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

func (self *Parser) parseMatch() *ast.Match {
	begin := self.expectNextIs(token.MATCH).Position
	value := self.mustExpr(self.parseOptionExpr(false))
	self.expectNextIs(token.LBR)
	var cases []ast.MatchCase
	other := util.None[*ast.Block]()
	for self.skipSEM(); !self.nextIs(token.RBR) && (self.skipNextIs(token.CASE) || self.expectNextIs(token.OTHER).Is(token.OTHER)); self.skipSEM() {
		caseBeginTok := self.curTok
		if caseBeginTok.Is(token.CASE) {
			name := self.expectNextIs(token.IDENT)
			var elems []ast.MatchCaseElem
			if self.skipNextIs(token.LPA) {
				elems = loopParseWithUtil(self, token.COM, token.RPA, func() ast.MatchCaseElem {
					mut := self.skipNextIs(token.MUT)
					pn := self.expectNextIs(token.IDENT)
					return ast.MatchCaseElem{
						Mutable: mut,
						Name:    pn,
					}
				})
				self.expectNextIs(token.RPA)
			}
			elemEnd := self.curTok.Position
			self.expectNextIs(token.COL)
			body := &ast.Block{
				Begin: caseBeginTok.Position,
				Stmts: self.parseStmtList(token.CASE, token.OTHER, token.RBR),
				End:   self.curTok.Position,
			}
			cases = append(cases, ast.MatchCase{
				Name:    name,
				Elems:   elems,
				ElemEnd: elemEnd,
				Body:    body,
			})
		} else {
			self.expectNextIs(token.COL)
			body := &ast.Block{
				Begin: caseBeginTok.Position,
				Stmts: self.parseStmtList(token.CASE, token.OTHER, token.RBR),
				End:   self.curTok.Position,
			}
			other = util.Some(body)
			break
		}
	}
	self.skipSEM()
	end := self.expectNextIs(token.RBR).Position
	return &ast.Match{
		Begin: begin,
		Value: value,
		Cases: cases,
		Other: other,
		End:   end,
	}
}
