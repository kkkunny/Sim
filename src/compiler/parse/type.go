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
	Pkgs        []*Package // 可能存在于的包，按顺序查找
	Name        lex.Token
	GenericArgs []Type
}

func NewTypeIdent(pkgs []*Package, name lex.Token, genericArgs []Type) *TypeIdent {
	return &TypeIdent{
		Pkgs:        pkgs,
		Name:        name,
		GenericArgs: genericArgs,
	}
}

func (self TypeIdent) Position() utils.Position {
	return self.Name.Pos
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

// TypeStruct 结构体类型
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

// TypeUnion 联合类型
type TypeUnion struct {
	Pos   utils.Position
	Elems []Type
}

func NewTypeUnion(pos utils.Position, elem ...Type) *TypeUnion {
	return &TypeUnion{
		Pos:   pos,
		Elems: elem,
	}
}

func (self TypeUnion) Position() utils.Position {
	return self.Pos
}

func (self TypeUnion) Type() {}

// ****************************************************************

// 类型或空
func (self *parser) parseTypeOrNil() Type {
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
	case lex.LT:
		return self.parseTypeUnion()
	case lex.STRUCT:
		return self.parseTypeStruct()
	default:
		return nil
	}
}

// 类型
func (self *parser) parseType() Type {
	typ := self.parseTypeOrNil()
	if typ == nil {
		self.throwErrorf(self.nextTok.Pos, "unknown type")
	}
	return typ
}

// 类型列表
func (self *parser) parseTypeList() (toks []Type) {
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
func (self *parser) parsParamTypeList() (toks []Type, varArg bool) {
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
func (self *parser) parseTypeIdent() *TypeIdent {
	// 语法解析
	var pkgPath *lex.Token
	name := self.expectNextIs(lex.IDENT)
	if self.skipNextIs(lex.CLL) {
		tmp := name
		pkgPath = &tmp
		name = self.expectNextIs(lex.IDENT)
	}
	genericArgs := self.parseGenericArgList()

	// 包解析
	if pkgPath != nil {
		pkg, ok := self.pkg.importMap[pkgPath.Source]
		if !ok {
			self.throwErrorf(pkgPath.Pos, "unknown package name")
		}
		return NewTypeIdent([]*Package{pkg}, name, genericArgs)
	} else {
		pkgs := make([]*Package, self.pkg.includeMap.Length()+1)
		pkgs[0] = self.pkg
		for iter := self.pkg.includeMap.Begin(); iter.HasValue(); iter.Next() {
			pkgs[self.pkg.includeMap.Length()-iter.Index()] = iter.Value()
		}
		return NewTypeIdent(pkgs, name, genericArgs)
	}
}

// 指针类型
func (self *parser) parseTypePtr() *TypePtr {
	begin := self.expectNextIs(lex.MUL).Pos
	elem := self.parseType()
	return NewTypePtr(utils.MixPosition(begin, elem.Position()), elem)
}

// 函数类型
func (self *parser) parseTypeFunc() *TypeFunc {
	begin := self.expectNextIs(lex.FUNC).Pos
	self.expectNextIs(lex.LPA)
	params, varArg := self.parsParamTypeList()
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	return NewTypeFunc(utils.MixPosition(begin, self.curTok.Pos), ret, params, varArg)
}

// 数组类型
func (self *parser) parseTypeArray() *TypeArray {
	begin := self.expectNextIs(lex.LBA).Pos
	size := self.parseIntExpr()
	self.expectNextIs(lex.RBA)
	elem := self.parseType()
	return NewTypeArray(utils.MixPosition(begin, elem.Position()), size, elem)
}

// 元组类型
func (self *parser) parseTypeTuple() *TypeTuple {
	begin := self.expectNextIs(lex.LPA).Pos
	elems := self.parseTypeList()
	end := self.expectNextIs(lex.RPA).Pos
	return NewTypeTuple(utils.MixPosition(begin, end), elems...)
}

// 结构体类型
func (self *parser) parseTypeStruct() *TypeStruct {
	begin := self.expectNextIs(lex.STRUCT).Pos
	self.expectNextIs(lex.LBR)
	var fields []types.Pair[bool, *NameAndType]
	for self.skipSem(); !self.nextIs(lex.RBR); {
		pub := self.skipNextIs(lex.PUB)
		fields = append(fields, types.NewPair(pub, self.parseNameAndType(false)))
		com := self.skipNextIs(lex.COM)
		self.skipSem()
		if !com {
			break
		}
	}
	end := self.expectNextIs(lex.RBR).Pos
	return NewTypeStruct(utils.MixPosition(begin, end), fields...)
}

// 联合类型
func (self *parser) parseTypeUnion() *TypeUnion {
	begin := self.expectNextIs(lex.LT).Pos
	elems := self.parseTypeList()
	end := self.expectNextIs(lex.GT).Pos
	return NewTypeUnion(utils.MixPosition(begin, end), elems...)
}
