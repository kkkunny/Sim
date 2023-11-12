package parse

import (
	. "github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseOptionType() util.Option[Type] {
	switch self.nextTok.Kind {
	case token.IDENT:
		return util.Some[Type](self.parseIdentType())
	case token.FUNC:
		return util.Some[Type](self.parseFuncType())
	case token.LBA:
		return util.Some[Type](self.parseArrayType())
	case token.LPA:
		return util.Some[Type](self.parseTupleType())
	default:
		return util.None[Type]()
	}
}

func (self *Parser) parseType() Type {
	t, ok := self.parseOptionType().Value()
	if !ok {
		errors.ThrowIllegalType(self.nextTok.Position)
	}
	return t
}

func (self *Parser) parseIdentType() *IdentType {
	return &IdentType{Name: self.expectNextIs(token.IDENT)}
}

func (self *Parser) parseFuncType() *FuncType {
	begin := self.expectNextIs(token.FUNC).Position
	self.expectNextIs(token.LPA)
	params := self.parseTypeList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	ret := self.parseOptionType()
	if v, ok := ret.Value(); ok {
		end = v.Position()
	}
	return &FuncType{
		Begin:  begin,
		Params: params,
		Ret:    ret,
		End:    end,
	}
}

func (self *Parser) parseArrayType() *ArrayType {
	begin := self.expectNextIs(token.LBA).Position
	size := self.expectNextIs(token.INTEGER)
	self.expectNextIs(token.RBA)
	elem := self.parseType()
	return &ArrayType{
		Begin: begin,
		Size:  size,
		Elem:  elem,
	}
}

func (self *Parser) parseTupleType() *TupleType {
	begin := self.expectNextIs(token.LPA).Position
	elems := self.parseTypeList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	return &TupleType{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseTypeList(end token.Kind) (res []Type) {
	return loopParseWithUtil(self, token.COM, end, func() Type {
		return self.parseType()
	})
}
