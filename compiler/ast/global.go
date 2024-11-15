package ast

import (
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Global 全局ast
type Global interface {
	Ast
	global()
}

// Import 包导入
type Import struct {
	Begin reader.Position
	Paths []token.Token
	Alias optional.Optional[token.Token]
}

func (self *Import) Position() reader.Position {
	if alias, ok := self.Alias.Value(); ok {
		return reader.MixPosition(self.Begin, alias.Position)
	}
	return reader.MixPosition(self.Begin, stlslices.Last(self.Paths).Position)
}

func (*Import) global() {}

type VariableDef interface {
	Global
	Stmt
	variable()
}

type VarDef struct {
	Mutable bool
	Name    token.Token
	Type    optional.Optional[Type]
}

// SingleVariableDef 单变量定义
type SingleVariableDef struct {
	Attrs  []Attr
	Begin  reader.Position
	Public bool
	Var    VarDef
	Value  optional.Optional[Expr]
}

func (self *SingleVariableDef) Position() reader.Position {
	if v, ok := self.Value.Value(); ok {
		return reader.MixPosition(self.Begin, v.Position())
	} else {
		return reader.MixPosition(self.Begin, self.Var.Type.MustValue().Position())
	}
}

func (*SingleVariableDef) stmt() {}

func (*SingleVariableDef) global() {}

func (*SingleVariableDef) variable() {}

// MultipleVariableDef 多变量定义
type MultipleVariableDef struct {
	Attrs  []Attr
	Begin  reader.Position
	Public bool
	Vars   []VarDef
	Value  optional.Optional[Expr]
	End    reader.Position
}

func (self *MultipleVariableDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*MultipleVariableDef) stmt() {}

func (*MultipleVariableDef) global() {}

func (*MultipleVariableDef) variable() {}

func (self *MultipleVariableDef) ToSingleList() []*SingleVariableDef {
	return lo.Map(self.Vars, func(item VarDef, _ int) *SingleVariableDef {
		return &SingleVariableDef{
			Attrs:  self.Attrs,
			Begin:  self.Begin,
			Public: self.Public,
			Var:    item,
			Value:  optional.None[Expr](),
		}
	})
}

// FuncDef 函数定义
type FuncDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	SelfType optional.Optional[token.Token]
	FuncDecl
	Body optional.Optional[*Block]
}

func (self *FuncDef) Position() reader.Position {
	if b, ok := self.Body.Value(); ok {
		return reader.MixPosition(self.Begin, b.Position())
	} else {
		return reader.MixPosition(self.Begin, self.FuncDecl.End)
	}
}

func (*FuncDef) global() {}

// TypeDef 类型定义
type TypeDef struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Target Type
}

func (self *TypeDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Target.Position())
}

func (*TypeDef) global() {}

// TypeAlias 类型别名
type TypeAlias struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Target Type
}

func (self *TypeAlias) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Target.Position())
}

func (*TypeAlias) global() {}

// Trait 特征
type Trait struct {
	Begin   reader.Position
	Public  bool
	Name    token.Token
	Methods []*FuncDecl
	End     reader.Position
}

func (self *Trait) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*Trait) global() {}
