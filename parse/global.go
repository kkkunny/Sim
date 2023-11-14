package parse

import (
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"

	. "github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseGlobal() Global {
	switch self.nextTok.Kind {
	case token.FUNC:
		return self.parseFuncDef()
	case token.STRUCT:
		return self.parseStructDef()
	case token.LET:
		return self.parseVariable()
	case token.IMPORT:
		return self.parseImport()
	default:
		errors.ThrowIllegalGlobal(self.nextTok.Position)
		return nil
	}
}

func (self *Parser) parseFuncDef() *FuncDef {
	begin := self.expectNextIs(token.FUNC).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LPA)
	args := loopParseWithUtil(self, token.COM, token.RPA, func() Param {
		mut := self.skipNextIs(token.MUT)
		pn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		pt := self.parseType()
		return Param{
			Mutable: mut,
			Name:    pn,
			Type:    pt,
		}
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

func (self *Parser) parseVariable() *Variable {
	begin := self.expectNextIs(token.LET).Position
	mut := self.skipNextIs(token.MUT)
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.COL)
	typ := self.parseType()
	self.expectNextIs(token.ASS)
	value := self.mustExpr(self.parseOptionExpr(true))
	return &Variable{
		Mutable: mut,
		Begin:   begin,
		Name:    name,
		Type:    typ,
		Value:   value,
	}
}

func (self *Parser) parseImport() *Import {
	begin := self.expectNextIs(token.IMPORT).Position
	var paths dynarray.DynArray[token.Token]
	for {
		paths.PushBack(self.expectNextIs(token.IDENT))
		if !self.skipNextIs(token.DOT) {
			break
		}
	}
	return &Import{
		Begin: begin,
		Paths: paths,
	}
}
