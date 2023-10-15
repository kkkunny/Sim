package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseGlobal() Global {
	switch self.nextTok.Kind {
	case token.FUNC:
		return self.parseFuncDef()
	default:
		// TODO: 编译时异常：未知的全局
		panic("unreachable")
	}
}

func (self *Parser) parseFuncDef() *FuncDef {
	begin := self.expectNextIs(token.FUNC).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LPA)
	self.expectNextIs(token.RPA)
	ret := util.Optional(self.parseTypeOrNil())
	body := self.parseBlock()
	return &FuncDef{
		Begin: begin,
		Name:  name,
		Ret:   ret,
		Body:  body,
	}
}
