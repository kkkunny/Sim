package parse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/reader"

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
		return self.parseFuncOrMethodDef(attrs, pub)
	case token.STRUCT:
		return self.parseStructDef(attrs, pub)
	case token.LET:
		return self.parseVariable(true, attrs, pub)
	case token.IMPORT:
		return self.parseImport(attrs)
	case token.TYPE:
		return self.parseTypeAlias(attrs, pub)
	default:
		errors.ThrowIllegalGlobal(self.nextTok.Position)
		return nil
	}
}

func (self *Parser) parseFuncOrMethodDef(attrs []ast.Attr, pub *token.Token) ast.Global {
	begin := self.expectNextIs(token.FUNC).Position
	if self.nextIs(token.LPA){
		return self.parseMethodDef(attrs, pub, begin)
	}else{
		return self.parseFuncDef(attrs, pub, begin)
	}
}

func (self *Parser) parseFuncDef(attrs []ast.Attr, pub *token.Token, begin reader.Position) ast.Global {
	expectAttrIn(attrs, new(ast.Extern), new(ast.NoReturn), new(ast.Inline), new(ast.NoInline))

	if pub != nil {
		begin = pub.Position
	}
	name := self.expectNextIs(token.IDENT)
	if self.nextIs(token.LT){
		return self.parseGenericFuncDef(attrs, begin, pub!=nil, name)
	}
	self.expectNextIs(token.LPA)
	params := loopParseWithUtil(self, token.COM, token.RPA, func() ast.Param {
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
	paramEnd := self.expectNextIs(token.RPA).Position
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
		Name:   name,
		Params: params,
		ParamEnd: paramEnd,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseMethodDef(attrs []ast.Attr, pub *token.Token, begin reader.Position) ast.Global {
	expectAttrIn(attrs, new(ast.NoReturn), new(ast.Inline), new(ast.NoInline))

	if pub != nil {
		begin = pub.Position
	}
	self.expectNextIs(token.LPA)
	mut := self.skipNextIs(token.MUT)
	scope := self.expectNextIs(token.IDENT)
	if self.nextIs(token.LT){
		return self.parseGenericStructMethodDef(attrs, begin, pub!=nil, mut, scope)
	}
	self.expectNextIs(token.RPA)
	name := self.expectNextIs(token.IDENT)
	if self.nextIs(token.LT){
		return self.parseGenericMethodDef(attrs, begin, pub!=nil, mut, scope, name)
	}
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
	body := self.parseBlock()
	return &ast.MethodDef{
		Attrs:  attrs,
		Begin:  begin,
		Public: pub != nil,
		ScopeMutable: mut,
		Scope: scope,
		Name:   name,
		Params: args,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseStructDef(attrs []ast.Attr, pub *token.Token) ast.Global {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.STRUCT).Position
	if pub != nil {
		begin = pub.Position
	}
	name := self.expectNextIs(token.IDENT)
	if self.nextIs(token.LT){
		return self.parseGenericStructDef(attrs, begin, pub!=nil, name)
	}
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() lo.Tuple3[bool, token.Token, ast.Type] {
		pub := self.skipNextIs(token.PUBLIC)
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		ft := self.parseType()
		return lo.Tuple3[bool, token.Token, ast.Type]{
			A: pub,
			B: fn,
			C: ft,
		}
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

func (self *Parser) parseVarDef(global bool) ast.VarDef {
	mut := self.skipNextIs(token.MUT)
	name := self.expectNextIs(token.IDENT)
	typ := util.None[ast.Type]()
	if self.nextIs(token.COL) || global{
		self.expectNextIs(token.COL)
		typ = util.Some(self.parseType())
	}
	return ast.VarDef{
		Mutable: mut,
		Name: name,
		Type: typ,
	}
}

func (self *Parser) parseVariable(global bool, attrs []ast.Attr, pub *token.Token) ast.VariableDef {
	begin := self.expectNextIs(token.LET).Position
	if pub != nil {
		begin = pub.Position
	}
	if !self.nextIs(token.LPA){
		return self.parseSingleVariable(begin, attrs, global, pub != nil)
	}else{
		return self.parseMultipleVariable(begin, attrs, global, pub != nil)
	}
}

func (self *Parser) parseSingleVariable(begin reader.Position, attrs []ast.Attr, global bool, pub bool) *ast.SingleVariableDef {
	expectAttrIn(attrs, new(ast.Extern))

	varDef := self.parseVarDef(global)
	value := util.None[ast.Expr]()
	if self.nextIs(token.ASS) || varDef.Type.IsNone(){
		self.expectNextIs(token.ASS)
		value = util.Some(self.mustExpr(self.parseOptionExpr(true)))
	}
	return &ast.SingleVariableDef{
		Attrs:   attrs,
		Public:  pub,
		Begin:   begin,
		Var: varDef,
		Value:   value,
	}
}

func (self *Parser) parseMultipleVariable(begin reader.Position, attrs []ast.Attr, global bool, pub bool) *ast.MultipleVariableDef {
	expectAttrIn(attrs)

	self.expectNextIs(token.LPA)
	var anyNoType bool
	varDefs := loopParseWithUtil(self, token.COM, token.RPA, func() ast.VarDef {
		varDef := self.parseVarDef(global)
		anyNoType = anyNoType || varDef.Type.IsNone()
		return varDef
	})
	self.expectNextIs(token.RPA)
	value := util.None[ast.Expr]()
	if self.nextIs(token.ASS) || anyNoType {
		self.expectNextIs(token.ASS)
		value = util.Some(self.mustExpr(self.parseOptionExpr(true)))
	}
	return &ast.MultipleVariableDef{
		Attrs:   attrs,
		Public:  pub,
		Begin:   begin,
		Vars: varDefs,
		Value:   value,
		End: self.curTok.Position,
	}
}

func (self *Parser) parseImport(attrs []ast.Attr) *ast.Import {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.IMPORT).Position
	var paths dynarray.DynArray[token.Token]
	for {
		paths.PushBack(self.expectNextIs(token.IDENT))
		if !self.skipNextIs(token.SCOPE) {
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

func (self *Parser) parseGenericFuncDef(attrs []ast.Attr, begin reader.Position, pub bool, name token.Token) *ast.GenericFuncDef {
	expectAttrIn(attrs, new(ast.NoReturn), new(ast.Inline), new(ast.NoInline))

	genericName := self.parseGenericNameDef(name)
	self.expectNextIs(token.LPA)
	params := loopParseWithUtil(self, token.COM, token.RPA, func() ast.Param {
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
	body := self.parseBlock()
	return &ast.GenericFuncDef{
		Attrs:  attrs,
		Begin:  begin,
		Public: pub,
		Name:   genericName,
		Params: params,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseGenericStructDef(attrs []ast.Attr, begin reader.Position, pub bool, name token.Token) *ast.GenericStructDef {
	expectAttrIn(attrs)

	genericName := self.parseGenericNameDef(name)
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() lo.Tuple3[bool, token.Token, ast.Type] {
		pub := self.skipNextIs(token.PUBLIC)
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		ft := self.parseType()
		return lo.Tuple3[bool, token.Token, ast.Type]{
			A: pub,
			B: fn,
			C: ft,
		}
	})
	end := self.expectNextIs(token.RBR).Position
	return &ast.GenericStructDef{
		Begin:  begin,
		Public: pub,
		Name:   genericName,
		Fields: fields,
		End:    end,
	}
}

func (self *Parser) parseGenericStructMethodDef(attrs []ast.Attr, begin reader.Position, pub bool, mut bool, scopeTok token.Token) ast.Global {
	scope := self.parseGenericNameDef(scopeTok)
	self.expectNextIs(token.RPA)
	nameTok := self.expectNextIs(token.IDENT)
	var name ast.GenericNameDef
	if self.nextIs(token.LT){
		name = self.parseGenericNameDef(nameTok)
	}else{
		name = ast.GenericNameDef{
			Name: nameTok,
			Params: ast.List[token.Token]{
				Begin: nameTok.Position,
				Data: nil,
				End: nameTok.Position,
			},
		}
	}
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
	body := self.parseBlock()
	return &ast.GenericStructMethodDef{
		Attrs:  attrs,
		Begin:  begin,
		Public: pub,
		ScopeMutable: mut,
		Scope: scope,
		Name:   name,
		Params: args,
		Ret:    ret,
		Body:   body,
	}
}

func (self *Parser) parseGenericMethodDef(attrs []ast.Attr, begin reader.Position, pub bool, mut bool, scope token.Token, nameTok token.Token) *ast.GenericMethodDef {
	name := self.parseGenericNameDef(nameTok)
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
	body := self.parseBlock()
	return &ast.GenericMethodDef{
		Attrs:  attrs,
		Begin:  begin,
		Public: pub,
		ScopeMutable: mut,
		Scope: scope,
		Name:   name,
		Params: args,
		Ret:    ret,
		Body:   body,
	}
}
