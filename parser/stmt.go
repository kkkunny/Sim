package parser

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseBlock() Block {
	begin := self.expectNextIs(token.LBR).Position
	end := self.expectNextIs(token.RBR).Position
	return Block{
		Begin: begin,
		End:   end,
	}
}
