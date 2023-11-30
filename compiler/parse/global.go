package parse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/ast"

	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseGlobal() ast.Global {
	attrs := self.parseAttrList()

	var pub *token.Token
	if self.skipNextIs(token.PUBLIC) {
		pub = &self.curTok
	}

	switch self.nextTok.Kind {
	case token.FUNC:
		return self.parseFuncDef(attrs, pub)
	case token.STRUCT:
		return self.parseStructDef(attrs, pub)
	case token.LET:
		return self.parseVariable(attrs, pub)
	case token.IMPORT:
		return self.parseImport(attrs)
	case token.TYPE:
		return self.parseTypeAlias(attrs, pub)
	default:
		errors.ThrowIllegalGlobal(self.nextTok.Position)
		return nil
	}
}

func (self *Parser) parseFuncDef(attrs []ast.Attr, pub *token.Token) *ast.FuncDef {
	expectAttrIn(attrs, new(ast.Extern))

	begin := self.expectNextIs(token.FUNC).Position
	if pub != nil {
		begin = pub.Position
	}
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LPA)
	args := loopParseWithUtil(self, token.COM, token.RPA, func() ast.Param {
		mut := self.skipNextIs(token.MUT)
		pn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		pt := self.parseType()
		return ast.Param{
			Mutable: mut,
			Name:    pn,
			Type:    pt,
		}
	})
	self.expectNextIs(token.RPA)
	ret := self.parseOptionType()
	body := stlbasic.TernaryAction(self.nextIs(token.LBR), func() util.Option[*ast.Block] {
		return util.Some(self.parseBlock())
	}, func() util.Option[*ast.Block] {
		return util.None[*ast.Block]()
	})
	return &ast.FuncDef{
		Attrs:  attrs,
		Begin:  begin,
		Public: pub != nil,
		Params: args,
		Name:   name,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseStructDef(attrs []ast.Attr, pub *token.Token) *ast.StructDef {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.STRUCT).Position
	if pub != nil {
		begin = pub.Position
	}
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() pair.Pair[token.Token, ast.Type] {
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		ft := self.parseType()
		return pair.NewPair(fn, ft)
	})
	end := self.expectNextIs(token.RBR).Position
	return &ast.StructDef{
		Begin:  begin,
		Public: pub != nil,
		Name:   name,
		Fields: fields,
		End:    end,
	}
}

func (self *Parser) parseVariable(attrs []ast.Attr, pub *token.Token) *ast.Variable {
	expectAttrIn(attrs, new(ast.Extern))

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
	return &ast.Variable{
		Attrs:   attrs,
		Public:  mut,
		Begin:   begin,
		Mutable: mut,
		Name:    name,
		Type:    typ,
		Value:   value,
	}
}

func (self *Parser) parseImport(attrs []ast.Attr) *ast.Import {
	expectAttrIn(attrs)

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
	return &ast.Import{
		Begin: begin,
		Paths: paths,
		Alias: alias,
	}
}

func (self *Parser) parseTypeAlias(attrs []ast.Attr, pub *token.Token) *ast.TypeAlias {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.TYPE).Position
	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.ASS)
	typ := self.parseType()
	return &ast.TypeAlias{
		Begin:  begin,
		Public: pub != nil,
		Name:   name,
		Type:   typ,
	}
}
