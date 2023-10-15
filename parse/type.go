package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseOptionType() util.Option[Type] {
	switch self.nextTok.Kind {
	case token.IDENT:
		return util.Some[Type](self.parseIdentType())
	case token.FUNC:
		return util.Some[Type](self.parseFuncType())
	default:
		return util.None[Type]()
	}
}

func (self *Parser) parseType() Type {
	t, ok := self.parseOptionType().Value()
	if !ok {
		// TODO: 编译时异常：期待一个类型
		panic("unreachable")
	}
	return t
}

func (self *Parser) parseIdentType() *IdentType {
	return &IdentType{Name: self.expectNextIs(token.IDENT)}
}

func (self *Parser) parseFuncType() *FuncType {
	begin := self.expectNextIs(token.FUNC).Position
	self.expectNextIs(token.LPA)
	end := self.expectNextIs(token.RPA).Position
	ret := self.parseOptionType()
	if v, ok := ret.Value(); ok {
		end = v.Position()
	}
	return &FuncType{
		Begin: begin,
		Ret:   ret,
		End:   end,
	}
}
