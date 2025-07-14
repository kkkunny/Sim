package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func (self *Parser) parseOptionType() optional.Optional[ast.Type] {
	switch self.nextTok.Kind {
	case token.IDENT:
		return optional.Some[ast.Type]((*ast.IdentType)(self.parseIdent()))
	case token.FUNC:
		return optional.Some[ast.Type](self.parseFuncType())
	case token.LBA:
		return optional.Some[ast.Type](self.parseArrayType())
	case token.LPA:
		return optional.Some[ast.Type](self.parseTupleOrLambdaType())
	case token.AND, token.LAND:
		return optional.Some[ast.Type](self.parseRefType())
	default:
		return optional.None[ast.Type]()
	}
}

func (self *Parser) parseType() ast.Type {
	t, ok := self.parseOptionType().Value()
	if !ok {
		errors.ThrowIllegalType(self.nextTok.Position)
	}
	return t
}

func (self *Parser) parseTypeInTypedef() ast.Type {
	if self.nextIs(token.STRUCT) {
		return self.parseStructType()
	} else if self.nextIs(token.ENUM) {
		return self.parseEnumType()
	}
	return self.parseType()
}

func (self *Parser) parseFuncType() *ast.FuncType {
	begin := self.expectNextIs(token.FUNC).Position
	self.expectNextIs(token.LPA)
	params := self.parseTypeList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	ret := self.parseOptionType()
	if v, ok := ret.Value(); ok {
		end = v.Position()
	}
	return &ast.FuncType{
		Begin:  begin,
		Params: params,
		Ret:    ret,
		End:    end,
	}
}

func (self *Parser) parseArrayType() *ast.ArrayType {
	begin := self.expectNextIs(token.LBA).Position
	size := self.expectNextIs(token.INTEGER)
	self.expectNextIs(token.RBA)
	elem := self.parseType()
	return &ast.ArrayType{
		Begin: begin,
		Size:  size,
		Elem:  elem,
	}
}

func (self *Parser) parseTupleOrLambdaType() ast.Type {
	begin := self.expectNextIs(token.LPA).Position
	elems := self.parseTypeList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	if !self.skipNextIs(token.ARROW) {
		return &ast.TupleType{
			Begin: begin,
			Elems: elems,
			End:   end,
		}
	}
	ret := self.parseType()
	return &ast.LambdaType{
		Begin:  begin,
		Params: elems,
		Ret:    ret,
	}
}

func (self *Parser) parseRefType() ast.Type {
	begin := self.expectNextIs(token.AND).Position
	mut := self.skipNextIs(token.MUT)
	return &ast.RefType{
		Begin: begin,
		Mut:   mut,
		Elem:  self.parseType(),
	}
}

func (self *Parser) parseStructType() *ast.StructType {
	begin := self.expectNextIs(token.STRUCT).Position
	self.expectNextIs(token.LBR)
	fields := self.parseFieldList(token.RBR)
	end := self.expectNextIs(token.RBR).Position
	return &ast.StructType{
		Begin:  begin,
		Fields: fields,
		End:    end,
	}
}

func (self *Parser) parseEnumType() *ast.EnumType {
	begin := self.expectNextIs(token.ENUM).Position
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() ast.EnumField {
		name := self.expectNextIs(token.IDENT)
		var elem optional.Optional[ast.Type]
		if self.skipNextIs(token.COL) {
			elem = optional.Some(self.parseType())
		}
		return ast.EnumField{
			Name: name,
			Elem: elem,
		}
	})
	end := self.expectNextIs(token.RBR).Position
	return &ast.EnumType{
		Begin:  begin,
		Fields: fields,
		End:    end,
	}
}
