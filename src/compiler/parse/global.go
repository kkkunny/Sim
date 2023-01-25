package parse

import (
	"fmt"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/types"
)

// Global 全局
type Global interface {
	Ast
	Global()
}

// Import 包导入
type Import struct {
	Pos      utils.Position
	Packages []lex.Token
	Suffix   *types.Either[bool, lex.Token] // 左为import ... as *， 右为import ... as alias
}

func NewImport(pos utils.Position, pkgs []lex.Token, suffix *types.Either[bool, lex.Token]) *Import {
	return &Import{
		Pos:      pos,
		Packages: pkgs,
		Suffix:   suffix,
	}
}

func (self Import) Position() utils.Position {
	return self.Pos
}

func (self Import) Global() {}

// TypeDef 类型定义
type TypeDef struct {
	Pos    utils.Position
	Public bool
	Name   lex.Token
	Impls  []Type
	Target Type
}

func NewTypeDef(pos utils.Position, pub bool, name lex.Token, impls []Type, target Type) *TypeDef {
	return &TypeDef{
		Pos:    pos,
		Public: pub,
		Name:   name,
		Impls:  impls,
		Target: target,
	}
}

func (self TypeDef) Position() utils.Position {
	return self.Pos
}

func (self TypeDef) Global() {}

// ExternFunction 外部函数声明
type ExternFunction struct {
	Pos    utils.Position
	Attrs  []Attr
	Public bool
	Ret    Type
	Name   lex.Token
	Params []*NameOrNilAndType
	VarArg bool
}

func NewExternFunction(pos utils.Position, attrs []Attr, pub bool, ret Type, name lex.Token, params []*NameOrNilAndType, varArg bool) *ExternFunction {
	return &ExternFunction{
		Pos:    pos,
		Attrs:  attrs,
		Public: pub,
		Ret:    ret,
		Name:   name,
		Params: params,
		VarArg: varArg,
	}
}

func (self ExternFunction) Position() utils.Position {
	return self.Pos
}

func (self ExternFunction) Global() {}

// Function 函数
type Function struct {
	Pos    utils.Position
	Attrs  []Attr
	Public bool
	Ret    Type
	Name   lex.Token
	Params []*NameOrNilAndType
	VarArg bool
	Body   *Block // 可能为空
}

func NewFunction(pos utils.Position, attrs []Attr, pub bool, ret Type, name lex.Token, params []*NameOrNilAndType, varArg bool, body *Block) *Function {
	return &Function{
		Pos:    pos,
		Attrs:  attrs,
		Public: pub,
		Ret:    ret,
		Name:   name,
		Params: params,
		VarArg: varArg,
		Body:   body,
	}
}

func (self Function) Position() utils.Position {
	return self.Pos
}

func (self Function) Global() {}

// Method 方法
type Method struct {
	Pos    utils.Position
	Attrs  []Attr
	Public bool
	Self   lex.Token
	Ret    Type
	Name   lex.Token
	Params []*NameOrNilAndType
	VarArg bool
	Body   *Block
}

func NewMethod(pos utils.Position, attrs []Attr, pub bool, self lex.Token, ret Type, name lex.Token, params []*NameOrNilAndType, varArg bool, body *Block) *Method {
	return &Method{
		Pos:    pos,
		Attrs:  attrs,
		Public: pub,
		Self:   self,
		Ret:    ret,
		Name:   name,
		Params: params,
		VarArg: varArg,
		Body:   body,
	}
}

func (self Method) Position() utils.Position {
	return self.Pos
}

func (self Method) Global() {}

// GlobalValue 全局变量
type GlobalValue struct {
	Attrs    []Attr
	Public   bool
	Variable *Variable
}

func NewGlobalValue(pos utils.Position, attrs []Attr, pub bool, t Type, name lex.Token, v Expr) *GlobalValue {
	return &GlobalValue{
		Attrs:    attrs,
		Public:   pub,
		Variable: NewVariable(pos, t, name, v),
	}
}

func (self GlobalValue) Position() utils.Position {
	return self.Variable.Pos
}

func (self GlobalValue) Global() {}

// ****************************************************************

var (
	errStrUnknownGlobal = "unknown global"
	errStrCanNotUseAttr = "can not use this attribute"
)

// 全局
func (self *Parser) parseGlobal() Global {
	var pub *lex.Token
	if self.skipNextIs(lex.PUB) {
		pub = &self.curTok
	}

	switch self.nextTok.Kind {
	case lex.IMPORT, lex.TYPE:
		return self.parseGlobalWithNoAttr(pub)
	case lex.Attr, lex.FUNC, lex.LET:
		return self.parseGlobalWithAttr(pub)
	default:
		fmt.Println(self.nextTok.Source)
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 全局（不带属性）
func (self *Parser) parseGlobalWithNoAttr(pub *lex.Token) Global {
	switch self.nextTok.Kind {
	case lex.IMPORT:
		if pub != nil {
			self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		}
		return self.parseImport()
	case lex.TYPE:
		return self.parseTypeDef(pub)
	default:
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 全局（带属性）
func (self *Parser) parseGlobalWithAttr(pub *lex.Token) Global {
	var attrs []Attr
	if pub == nil {
		for self.nextIs(lex.Attr) {
			if len(attrs) == 0 && pub != nil {
				self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
			}
			attrs = append(attrs, self.parseAttr())
			self.expectNextIs(lex.SEM)
		}

		if self.skipNextIs(lex.PUB) {
			pub = &self.curTok
		}
	}

	switch self.nextTok.Kind {
	case lex.FUNC:
		return self.parseFunction(pub, attrs)
	case lex.LET:
		return self.parseGlobalValue(pub, attrs)
	default:
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 包导入
func (self *Parser) parseImport() *Import {
	begin := self.expectNextIs(lex.IMPORT).Pos
	pkgs := self.parseTokenListAtLeastOne(lex.DOT)
	var suffix *types.Either[bool, lex.Token]
	if self.skipNextIs(lex.AS) {
		var either types.Either[bool, lex.Token]
		if self.skipNextIs(lex.MUL) {
			either = types.Left[bool, lex.Token](true)
		} else {
			either = types.Right[bool, lex.Token](self.expectNextIs(lex.IDENT))
		}
		suffix = &either
	}
	return NewImport(utils.MixPosition(begin, pkgs[len(pkgs)-1].Pos), pkgs, suffix)
}

// 类型定义
func (self *Parser) parseTypeDef(pub *lex.Token) *TypeDef {
	begin := self.expectNextIs(lex.TYPE).Pos
	name := self.expectNextIs(lex.IDENT)
	var impls []Type
	if self.skipNextIs(lex.LPA) {
		for {
			impls = append(impls, self.parseType())
			if !self.skipNextIs(lex.COM) {
				break
			}
		}
		self.expectNextIs(lex.RPA)
	}
	target := self.parseType()
	return NewTypeDef(utils.MixPosition(begin, target.Position()), pub != nil, name, impls, target)
}

// 函数
func (self *Parser) parseFunction(pub *lex.Token, attrs []Attr) Global {
	var isExtern bool
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrExtern:
			isExtern = true
		case *AttrLink, *AttrNoReturn, *AttrInline, *AttrInit, *AttrFini:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}

	self.expectNextIs(lex.FUNC)
	var begin utils.Position
	if len(attrs) > 0 {
		begin = attrs[0].Position()
	} else if pub != nil {
		begin = pub.Pos
	} else {
		begin = self.curTok.Pos
	}

	if self.nextIs(lex.LPA) {
		return self.parseMethod(begin, pub != nil, attrs)
	}

	name := self.expectNextIs(lex.IDENT)
	self.expectNextIs(lex.LPA)
	params, varArg := self.parseParamList()
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	var body *Block
	if !isExtern || self.nextIs(lex.LBR) {
		body = self.parseBlock()
	}

	pos := utils.MixPosition(begin, self.curTok.Pos)
	if len(params) != 0 || ret != nil {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrInit, *AttrFini:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
	}
	if isExtern && body == nil {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrExtern, *AttrLink, *AttrNoReturn, *AttrInit, *AttrFini:
			default:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
		return NewExternFunction(pos, attrs, pub != nil, ret, name, params, varArg)
	} else {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrExtern, *AttrNoReturn, *AttrInline, *AttrInit, *AttrFini:
			default:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
		return NewFunction(pos, attrs, pub != nil, ret, name, params, varArg, body)
	}
}

// 方法
func (self *Parser) parseMethod(begin utils.Position, pub bool, attrs []Attr) *Method {
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrNoReturn, *AttrInline:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}

	self.expectNextIs(lex.LPA)
	selfTok := self.expectNextIs(lex.IDENT)
	self.expectNextIs(lex.RPA)

	name := self.expectNextIs(lex.IDENT)
	self.expectNextIs(lex.LPA)
	params, varArg := self.parseParamList()
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	var body *Block
	if self.nextIs(lex.LBR) {
		body = self.parseBlock()
	}
	return NewMethod(utils.MixPosition(begin, self.curTok.Pos), attrs, pub, selfTok, ret, name, params, varArg, body)
}

// 全局变量
func (self *Parser) parseGlobalValue(pub *lex.Token, attrs []Attr) *GlobalValue {
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrExtern, *AttrLink:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}
	v := self.parseVariable()
	var begin utils.Position
	if len(attrs) > 0 {
		begin = attrs[0].Position()
	} else if pub != nil {
		begin = pub.Pos
	} else {
		begin = v.Pos
	}
	return NewGlobalValue(utils.MixPosition(begin, v.Position()), attrs, pub != nil, v.Type, v.Name, v.Value)
}
