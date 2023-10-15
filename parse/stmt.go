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
	default:
		// TODO: 编译时异常：未知的语句
		panic("unreachable")
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
	return &Return{Token: self.expectNextIs(token.RETURN)}
}
