package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseStmt() Stmt {
	// TODO: stmt
	panic("unreachable")
}

func (self *Parser) parseBlock() *Block {
	begin := self.expectNextIs(token.LBR).Position
	self.skipSEM()
	end := self.expectNextIs(token.RBR).Position
	return &Block{
		Begin: begin,
		End:   end,
	}
}
