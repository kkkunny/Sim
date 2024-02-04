package ast

import (
	"github.com/kkkunny/stl/container/dynarray"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

// Type 类型
type Type interface {
	Ast
	typ()
}

// IdentType 标识符类型
type IdentType Ident

func (self *IdentType) Position() reader.Position {
	return (*Ident)(self).Position()
}

func (self *IdentType) typ() {}

// FuncType 函数类型
type FuncType struct {
	Begin  reader.Position
	Params []Type
	Ret    util.Option[Type]
	End    reader.Position
}

func (self *FuncType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *FuncType) typ() {}

// ArrayType 数组类型
type ArrayType struct {
	Begin reader.Position
	Size  token.Token
	Elem  Type
}

func (self *ArrayType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Elem.Position())
}

func (self *ArrayType) typ() {}

// TupleType 元组类型
type TupleType struct {
	Begin reader.Position
	Elems []Type
	End   reader.Position
}

func (self *TupleType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *TupleType) typ() {}

// UnionType 联合类型
type UnionType struct {
	Begin reader.Position
	Elems dynarray.DynArray[Type]
	End   reader.Position
}

func (self *UnionType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *UnionType) typ() {}

// RefType 引用类型
type RefType struct {
	Begin reader.Position
	Mut bool
	Elem  Type
}

func (self *RefType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Elem.Position())
}

func (self *RefType) typ() {}

type SelfType struct {
	Token token.Token
}

func (self *SelfType) Position() reader.Position {
	return self.Token.Position
}

func (self *SelfType) typ() {}
