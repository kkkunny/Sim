package parse

import (
	"github.com/kkkunny/stl/container/optional"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/reader"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
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
	case token.LET:
		return self.parseVariable(true, attrs, pub)
	case token.IMPORT:
		return self.parseImport(attrs)
	case token.TYPE:
		return self.parseTypeDefOrAlias(attrs, pub)
	case token.TRAIT:
		return self.parseTrait(attrs, pub)
	default:
		errors.ThrowIllegalGlobal(self.nextTok.Position)
		return nil
	}
}

func (self *Parser) parseFuncDef(attrs []ast.Attr, pub *token.Token) ast.Global {
	expectAttrIn(attrs, new(ast.Extern), new(ast.Inline), new(ast.NoInline), new(ast.VarArg))

	var selfType optional.Optional[token.Token]
	beforeNameFn := func() {
		if self.skipNextIs(token.LPA) {
			selfType = optional.Some(self.expectNextIs(token.IDENT))
			self.expectNextIs(token.RPA)
		}
	}

	var genericParams optional.Optional[*ast.GenericParamList]
	afterNameFn := func() {
		if selfType.IsNone() {
			genericParams = self.parseGenericParamList()
		}
		if genericParams.IsSome() {
			expectAttrIn(attrs, new(ast.Inline), new(ast.NoInline))
		}
	}
	decl := self.parseFuncDecl(beforeNameFn, afterNameFn)

	begin := stlval.TernaryAction(pub == nil, func() reader.Position {
		return decl.Begin
	}, func() reader.Position {
		return pub.Position
	})
	body := stlval.TernaryAction(self.nextIs(token.LBR), func() optional.Optional[*ast.Block] {
		return optional.Some(self.parseBlock())
	}, func() optional.Optional[*ast.Block] {
		return optional.None[*ast.Block]()
	})
	return &ast.FuncDef{
		Attrs:         attrs,
		Begin:         begin,
		Public:        pub != nil,
		SelfType:      selfType,
		FuncDecl:      decl,
		GenericParams: genericParams,
		Body:          body,
	}
}

func (self *Parser) parseVarDef(global bool) ast.VarDef {
	mut := self.skipNextIs(token.MUT)
	name := self.expectNextIs(token.IDENT)
	typ := optional.None[ast.Type]()
	if self.nextIs(token.COL) || global {
		self.expectNextIs(token.COL)
		typ = optional.Some(self.parseType())
	}
	return ast.VarDef{
		Mutable: mut,
		Name:    name,
		Type:    typ,
	}
}

func (self *Parser) parseVariable(global bool, attrs []ast.Attr, pub *token.Token) ast.VariableDef {
	begin := self.expectNextIs(token.LET).Position
	if pub != nil {
		begin = pub.Position
	}
	if !self.nextIs(token.LPA) {
		return self.parseSingleVariable(begin, attrs, global, pub != nil)
	} else {
		return self.parseMultipleVariable(begin, attrs, global, pub != nil)
	}
}

func (self *Parser) parseSingleVariable(begin reader.Position, attrs []ast.Attr, global bool, pub bool) *ast.SingleVariableDef {
	expectAttrIn(attrs, new(ast.Extern))

	varDef := self.parseVarDef(global)
	value := optional.None[ast.Expr]()
	if self.nextIs(token.ASS) || varDef.Type.IsNone() {
		self.expectNextIs(token.ASS)
		value = optional.Some(self.mustExpr(self.parseOptionExpr(true)))
	}
	return &ast.SingleVariableDef{
		Attrs:  attrs,
		Public: pub,
		Begin:  begin,
		Var:    varDef,
		Value:  value,
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
	value := optional.None[ast.Expr]()
	if self.nextIs(token.ASS) || anyNoType {
		self.expectNextIs(token.ASS)
		value = optional.Some(self.mustExpr(self.parseOptionExpr(true)))
	}
	return &ast.MultipleVariableDef{
		Attrs:  attrs,
		Public: pub,
		Begin:  begin,
		Vars:   varDefs,
		Value:  value,
		End:    self.curTok.Position,
	}
}

func (self *Parser) parseImport(attrs []ast.Attr) *ast.Import {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.IMPORT).Position
	var paths []token.Token
	for {
		paths = append(paths, self.expectNextIs(token.IDENT))
		if !self.skipNextIs(token.SCOPE) {
			break
		}
	}
	alias := optional.None[token.Token]()
	if self.skipNextIs(token.AS) {
		if self.skipNextIs(token.MUL) {
			alias = optional.Some(self.curTok)
		} else {
			alias = optional.Some(self.expectNextIs(token.IDENT))
		}
	}
	return &ast.Import{
		Begin: begin,
		Paths: paths,
		Alias: alias,
	}
}

func (self *Parser) parseTypeDefOrAlias(attrs []ast.Attr, pub *token.Token) ast.Global {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.TYPE).Position
	if pub != nil {
		begin = pub.Position
	}
	name := self.expectNextIs(token.IDENT)
	genericParams := self.parseGenericParamList()

	var isAlias bool
	if genericParams.IsNone() {
		isAlias = self.skipNextIs(token.ASS)
	}
	if isAlias {
		return &ast.TypeAlias{
			Begin:  begin,
			Public: pub != nil,
			Name:   name,
			Target: self.parseType(),
		}
	} else {
		return &ast.TypeDef{
			Begin:         begin,
			Public:        pub != nil,
			Name:          name,
			GenericParams: genericParams,
			Target:        self.parseTypeInTypedef(),
		}
	}
}

func (self *Parser) parseTrait(attrs []ast.Attr, pub *token.Token) *ast.Trait {
	expectAttrIn(attrs)

	begin := self.expectNextIs(token.TRAIT).Position
	if pub != nil {
		begin = pub.Position
	}

	name := self.expectNextIs(token.IDENT)
	self.expectNextIs(token.LBR)
	methods := loopParseWithUtil(self, token.COM, token.RBR, func() *ast.FuncDecl {
		return stlval.Ptr(self.parseFuncDecl(nil, nil))
	})
	end := self.expectNextIs(token.RBR).Position

	return &ast.Trait{
		Begin:   begin,
		Public:  pub != nil,
		Name:    name,
		Methods: methods,
		End:     end,
	}
}
