package parse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"

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

func (self *Parser) parseFuncDef(attrs []ast.Attr, pub *token.Token, begin reader.Position) *ast.FuncDef {
	if pub != nil {
		begin = pub.Position
	}
	name := self.parseGenericNameDef(self.expectNextIs(token.IDENT))
	if name.Params.IsNone(){
		expectAttrIn(attrs, new(ast.Extern), new(ast.NoReturn), new(ast.Inline), new(ast.NoInline), new(ast.VarArg))
	}else{
		expectAttrIn(attrs, new(ast.NoReturn), new(ast.Inline), new(ast.NoInline))
	}
	self.expectNextIs(token.LPA)
	params := self.parseParamList(token.RPA)
	paramEnd := self.expectNextIs(token.RPA).Position
	ret := self.parseOptionType()
	body := stlbasic.TernaryAction(name.Params.IsSome() || self.nextIs(token.LBR), func() util.Option[*ast.Block] {
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

func (self *Parser) parseMethodDef(attrs []ast.Attr, pub *token.Token, begin reader.Position) *ast.MethodDef {
	expectAttrIn(attrs, new(ast.NoReturn), new(ast.Inline), new(ast.NoInline))

	if pub != nil {
		begin = pub.Position
	}
	self.expectNextIs(token.LPA)
	scope := self.parseGenericNameDef(self.expectNextIs(token.IDENT))
	self.expectNextIs(token.RPA)
	name := self.parseGenericNameDef(self.expectNextIs(token.IDENT))
	self.expectNextIs(token.LPA)
	params := self.parseParamList(token.RPA)
	self.expectNextIs(token.RPA)
	ret := self.parseOptionType()
	body := self.parseBlock()
	return &ast.MethodDef{
		Attrs:    attrs,
		Begin:    begin,
		Public:   pub != nil,
		SelfType: scope,
		Name:     name,
		Params:   params,
		Ret:      ret,
		Body:     body,
	}
}

func (self *Parser) parseStructDef(attrs []ast.Attr, pub *token.Token) *ast.StructDef {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.STRUCT).Position
	if pub != nil {
		begin = pub.Position
	}
	name := self.parseGenericNameDef(self.expectNextIs(token.IDENT))
	self.expectNextIs(token.LBR)
	fields := self.parseFieldList(token.RBR)
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
	}, true)
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
