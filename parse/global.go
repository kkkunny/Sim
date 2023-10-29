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
	case token.STRUCT:
		return self.parseStructDef()
	default:
		// TODO: 编译时异常：未知的全局
		panic("编译时异常：未知的全局")
	}
}

func (self *Parser) parseFuncDef() *FuncDef {
	begin := self.expectNextIs(token.FUNC).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LPA)
	args := loopParseWithUtil(self, token.COM, token.RPA, func() pair.Pair[token.Token, Type] {
		pn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		pt := self.parseType()
		return pair.NewPair(pn, pt)
	})
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

func (self *Parser) parseStructDef() *StructDef {
	begin := self.expectNextIs(token.STRUCT).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() pair.Pair[token.Token, Type] {
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		ft := self.parseType()
		return pair.NewPair(fn, ft)
	})
	end := self.expectNextIs(token.RBR).Position
	return &StructDef{
		Begin:  begin,
		Name:   name,
		Fields: fields,
		End:    end,
	}
}
