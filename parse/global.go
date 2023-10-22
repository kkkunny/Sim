package parse

import (
	"github.com/kkkunny/stl/container/pair"

	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
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
	var args []pair.Pair[token.Token, Type]
	for self.skipSEM(); !self.nextIs(token.RPA); self.skipSEM() {
		pn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		pt := self.parseType()
		args = append(args, pair.NewPair(pn, pt))
		if !self.skipNextIs(token.COM) {
			break
		}
	}
	self.expectNextIs(token.RPA)
	ret := self.parseOptionType()
	body := self.parseBlock()
	return &FuncDef{
		Begin:  begin,
		Params: args,
		Name:   name,
		Ret:    ret,
		Body:   body,
	}
}
