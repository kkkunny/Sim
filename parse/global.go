package parse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"

	. "github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseGlobal() Global {
	var pub *token.Token
	if self.skipNextIs(token.PUBLIC) {
		pub = &self.curTok
	}

	switch self.nextTok.Kind {
	case token.FUNC:
		return self.parseFuncDef(pub)
	case token.STRUCT:
		return self.parseStructDef(pub)
	case token.LET:
		return self.parseVariable(pub)
	case token.IMPORT:
		return self.parseImport()
	default:
		errors.ThrowIllegalGlobal(self.nextTok.Position)
		return nil
	}
}

func (self *Parser) parseFuncDef(pub *token.Token) *FuncDef {
	begin := self.expectNextIs(token.FUNC).Position
	if pub != nil {
		begin = pub.Position
	}
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
	body := stlbasic.TernaryAction(self.nextIs(token.LBR), func() util.Option[*Block] {
		return util.Some(self.parseBlock())
	}, func() util.Option[*Block] {
		return util.None[*Block]()
	})
	return &FuncDef{
		Begin:  begin,
		Public: pub != nil,
		Params: args,
		Name:   name,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseStructDef(pub *token.Token) *StructDef {
	begin := self.expectNextIs(token.STRUCT).Position
	if pub != nil {
		begin = pub.Position
	}
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
		Public: pub != nil,
		Name:   name,
		Fields: fields,
		End:    end,
	}
}

func (self *Parser) parseVariable(pub *token.Token) *Variable {
	begin := self.expectNextIs(token.LET).Position
	if pub != nil {
		begin = pub.Position
	}
	mut := self.skipNextIs(token.MUT)
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.COL)
	typ := self.parseType()
	self.expectNextIs(token.ASS)
	value := self.mustExpr(self.parseOptionExpr(true))
	return &Variable{
		Public:  mut,
		Begin:   begin,
		Mutable: mut,
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
	alias := util.None[token.Token]()
	if self.skipNextIs(token.AS) {
		if self.skipNextIs(token.MUL) {
			alias = util.Some(self.curTok)
		} else {
			alias = util.Some(self.expectNextIs(token.IDENT))
		}
	}
	return &Import{
		Begin: begin,
		Paths: paths,
		Alias: alias,
	}
}
