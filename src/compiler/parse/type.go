package parse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/types"
)

// Type 类型
type Type interface {
	Ast
	Type()
}

// TypeIdent 标识符类型
type TypeIdent struct {
	Pkg  *lex.Token
	Name lex.Token
}

func NewTypeIdent(pkg *lex.Token, name lex.Token) *TypeIdent {
	return &TypeIdent{
		Pkg:  pkg,
		Name: name,
	}
}

func (self TypeIdent) Position() utils.Position {
	if self.Pkg == nil {
		return self.Name.Pos
	} else {
		return utils.MixPosition(self.Pkg.Pos, self.Name.Pos)
	}
}

func (self TypeIdent) Type() {}

// TypePtr 指针类型
type TypePtr struct {
	Pos  utils.Position
	Elem Type
}

func NewTypePtr(pos utils.Position, elem Type) *TypePtr {
	return &TypePtr{
		Pos:  pos,
		Elem: elem,
	}
}

func (self TypePtr) Position() utils.Position {
	return self.Pos
}

func (self TypePtr) Type() {}

// TypeFunc 函数类型
type TypeFunc struct {
	Pos    utils.Position
	Ret    Type // 可能为空
	Params []Type
	VarArg bool
}

func NewTypeFunc(pos utils.Position, ret Type, params []Type, varArg bool) *TypeFunc {
	return &TypeFunc{
		Pos:    pos,
		Ret:    ret,
		Params: params,
		VarArg: varArg,
	}
}

func (self TypeFunc) Position() utils.Position {
	return self.Pos
}

func (self TypeFunc) Type() {}

// TypeArray 数组类型
type TypeArray struct {
	Pos  utils.Position
	Size *Int
	Elem Type
}

func NewTypeArray(pos utils.Position, size *Int, elem Type) *TypeArray {
	return &TypeArray{
		Pos:  pos,
		Size: size,
		Elem: elem,
	}
}

func (self TypeArray) Position() utils.Position {
	return self.Pos
}

func (self TypeArray) Type() {}

// TypeTuple 元组类型
type TypeTuple struct {
	Pos   utils.Position
	Elems []Type
}

func NewTypeTuple(pos utils.Position, elem ...Type) *TypeTuple {
	return &TypeTuple{
		Pos:   pos,
		Elems: elem,
	}
}

func (self TypeTuple) Position() utils.Position {
	return self.Pos
}

func (self TypeTuple) Type() {}

// TypeStruct 元组类型
type TypeStruct struct {
	Pos    utils.Position
	Fields []types.Pair[bool, *NameAndType]
}

func NewTypeStruct(pos utils.Position, field ...types.Pair[bool, *NameAndType]) *TypeStruct {
	return &TypeStruct{
		Pos:    pos,
		Fields: field,
	}
}

func (self TypeStruct) Position() utils.Position {
	return self.Pos
}

func (self TypeStruct) Type() {}

// TypeInterface 接口类型
type TypeInterface struct {
	Pos    utils.Position
	Fields []*NameAndType
}

func NewTypeInterface(pos utils.Position, field ...*NameAndType) *TypeInterface {
	return &TypeInterface{
		Pos:    pos,
		Fields: field,
	}
}

func (self TypeInterface) Position() utils.Position {
	return self.Pos
}

func (self TypeInterface) Type() {}

// ****************************************************************

// 类型或空
func (self *Parser) parseTypeOrNil() Type {
	switch self.nextTok.Kind {
	case lex.IDENT:
		return self.parseTypeIdent()
	case lex.MUL:
		return self.parseTypePtr()
	case lex.FUNC:
		return self.parseTypeFunc()
	case lex.LBA:
		return self.parseTypeArray()
	case lex.LPA:
		return self.parseTypeTuple()
	case lex.STRUCT:
		return self.parseTypeStruct()
	case lex.INTERFACE:
		return self.parseTypeInterface()
	default:
		return nil
	}
}

// 类型
func (self *Parser) parseType() Type {
	typ := self.parseTypeOrNil()
	if typ == nil {
		self.throwErrorf(self.nextTok.Pos, "unknown type")
	}
	return typ
}

// 类型列表
func (self *Parser) parseTypeList() (toks []Type) {
	for {
		if len(toks) == 0 {
			typ := self.parseTypeOrNil()
			if typ == nil {
				break
			}
			toks = append(toks, typ)
		} else {
			toks = append(toks, self.parseType())
		}
		if !self.skipNextIs(lex.COM) {
			break
		}
	}
	return toks
}

// 函数参数类型列表
func (self *Parser) parsParamTypeList() (toks []Type, varArg bool) {
	for {
		if self.skipNextIs(lex.ELL) {
			varArg = true
			break
		} else if len(toks) == 0 {
			typ := self.parseTypeOrNil()
			if typ == nil {
				break
			}
			toks = append(toks, typ)
		} else {
			toks = append(toks, self.parseType())
		}
		if !self.skipNextIs(lex.COM) {
			break
		}
	}
	return toks, varArg
}

// 标识符类型
func (self *Parser) parseTypeIdent() *TypeIdent {
	pkg := self.expectNextIs(lex.IDENT)
	if self.skipNextIs(lex.CLL) {
		name := self.expectNextIs(lex.IDENT)
		return NewTypeIdent(&pkg, name)
	}
	return NewTypeIdent(nil, pkg)
}

// 指针类型
func (self *Parser) parseTypePtr() *TypePtr {
	begin := self.expectNextIs(lex.MUL).Pos
	elem := self.parseType()
	return NewTypePtr(utils.MixPosition(begin, elem.Position()), elem)
}

// 函数类型
func (self *Parser) parseTypeFunc() *TypeFunc {
	begin := self.expectNextIs(lex.FUNC).Pos
	self.expectNextIs(lex.LPA)
	params, varArg := self.parsParamTypeList()
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	return NewTypeFunc(utils.MixPosition(begin, self.curTok.Pos), ret, params, varArg)
}

// 数组类型
func (self *Parser) parseTypeArray() *TypeArray {
	begin := self.expectNextIs(lex.LBA).Pos
	size := self.parseIntExpr()
	self.expectNextIs(lex.RBA)
	elem := self.parseType()
	return NewTypeArray(utils.MixPosition(begin, elem.Position()), size, elem)
}

// 元组类型
func (self *Parser) parseTypeTuple() *TypeTuple {
	begin := self.expectNextIs(lex.LPA).Pos
	elems := self.parseTypeList()
	end := self.expectNextIs(lex.RPA).Pos
	return NewTypeTuple(utils.MixPosition(begin, end), elems...)
}

// 结构体类型
func (self *Parser) parseTypeStruct() *TypeStruct {
	begin := self.expectNextIs(lex.STRUCT).Pos
	self.expectNextIs(lex.LBR)
	var fields []types.Pair[bool, *NameAndType]
	for self.skipSem(); !self.nextIs(lex.RBR); self.skipSem() {
		pub := self.skipNextIs(lex.PUB)
		fields = append(fields, types.NewPair(pub, self.parseNameAndType(false)))
		self.expectNextIs(lex.SEM)
	}
	end := self.expectNextIs(lex.RBR).Pos
	return NewTypeStruct(utils.MixPosition(begin, end), fields...)
}

// 接口类型
func (self *Parser) parseTypeInterface() *TypeInterface {
	begin := self.expectNextIs(lex.INTERFACE).Pos
	self.expectNextIs(lex.LBR)
	var fields []*NameAndType
	for self.skipSem(); !self.nextIs(lex.RBR); self.skipSem() {
		field := self.parseNameAndType(true)
		fields = append(fields, field)
		self.expectNextIs(lex.SEM)
	}
	end := self.expectNextIs(lex.RBR).Pos
	return NewTypeInterface(utils.MixPosition(begin, end), fields...)
}
