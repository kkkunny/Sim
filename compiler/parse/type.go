package parse

import (
	"github.com/kkkunny/stl/container/dynarray"

	"github.com/kkkunny/Sim/ast"

	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseOptionType() util.Option[ast.Type] {
	switch self.nextTok.Kind {
	case token.IDENT:
		return util.Some[ast.Type](self.parseIdentType())
	case token.FUNC:
		return util.Some[ast.Type](self.parseFuncType())
	case token.LBA:
		return util.Some[ast.Type](self.parseArrayType())
	case token.LPA:
		return util.Some[ast.Type](self.parseTupleType())
	case token.LT:
		return util.Some[ast.Type](self.parseUnionType())
	case token.MUL:
		return util.Some[ast.Type](self.parsePtrOrRefType())
	case token.SELFTYPE:
		return util.Some[ast.Type](self.parseSelfType())
	default:
		return util.None[ast.Type]()
	}
}

func (self *Parser) parseType() ast.Type {
	t, ok := self.parseOptionType().Value()
	if !ok {
		errors.ThrowIllegalType(self.nextTok.Position)
	}
	return t
}

func (self *Parser) parseIdentType() *ast.IdentType {
	var pkg util.Option[token.Token]
	var name ast.GenericName
	pkgOrName := self.expectNextIs(token.IDENT)
	if !self.skipNextIs(token.SCOPE) {
		pkg, name = util.None[token.Token](), self.parseGenericName(pkgOrName, true)
	} else if !self.nextIs(token.LT){
		pkg, name = util.Some(pkgOrName), self.parseGenericName(self.expectNextIs(token.IDENT))
	} else {
		pkg, name = util.None[token.Token](), self.parseGenericName(pkgOrName)
	}
	return &ast.IdentType{
		Pkg:  pkg,
		Name: name,
	}
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

func (self *Parser) parseTupleType() *ast.TupleType {
	begin := self.expectNextIs(token.LPA).Position
	elems := self.parseTypeList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	return &ast.TupleType{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseUnionType() *ast.UnionType {
	begin := self.expectNextIs(token.LT).Position
	elems := dynarray.NewDynArrayWith[ast.Type](self.parseTypeList(token.GT, true)...)
	end := self.expectNextIs(token.GT).Position
	return &ast.UnionType{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parsePtrOrRefType() ast.Type {
	begin := self.expectNextIs(token.MUL).Position
	if self.skipNextIs(token.QUE){
		elem := self.parseType()
		return &ast.PtrType{
			Begin: begin,
			Elem:  elem,
		}
	}else{
		return &ast.RefType{
			Begin: begin,
			Elem:  self.parseType(),
		}
	}
}

func (self *Parser) parseSelfType()*ast.SelfType{
	return &ast.SelfType{Token: self.expectNextIs(token.SELFTYPE)}
}
