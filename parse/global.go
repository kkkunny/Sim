package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseGlobal() Global {
	switch self.nextTok.Kind {
	case token.FUNC:
		return self.parseFuncDef()
	default:
		// TODO: 报错
		panic("unreachable")
	}
}

func (self *Parser) parseFuncDef() *FuncDef {
	begin := self.expectNextIs(token.FUNC).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LPA)
	self.expectNextIs(token.RPA)
	body := self.parseBlock()
	return &FuncDef{
		Begin: begin,
		Name:  name,
		Body:  body,
	}
}
