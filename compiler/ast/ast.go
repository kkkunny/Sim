package ast

import (
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

// Ast 抽象语法树
type Ast interface {
	Position() reader.Position
}

type Param struct {
	Mutable util.Option[token.Token]
	Name    util.Option[token.Token]
	Type    Type
}

func (self Param) Position() reader.Position {
	if self.Mutable.IsSome() {
		return reader.MixPosition(self.Mutable.MustValue().Position, self.Type.Position())
	} else if self.Name.IsSome() {
		return reader.MixPosition(self.Name.MustValue().Position, self.Type.Position())
	} else {
		return self.Type.Position()
	}
}

type Field struct {
	Public  bool
	Mutable bool
	Name    token.Token
	Type    Type
}

type List[T any] struct {
	Begin reader.Position
	Data  []T
	End   reader.Position
}

func (self List[T]) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

// Ident 标识符
type Ident struct {
	Pkg  util.Option[token.Token]
	Name token.Token
}

func (self *Ident) Position() reader.Position {
	if pkg, ok := self.Pkg.Value(); ok {
		return reader.MixPosition(pkg.Position, self.Name.Position)
	}
	return self.Name.Position
}

type GenericParamList struct {
	Begin  reader.Position
	Params []token.Token
	End    reader.Position
}

func (self *GenericParamList) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

// FuncDecl 函数声明
type FuncDecl struct {
	Begin  reader.Position
	Name   token.Token
	Params []Param
	Ret    util.Option[Type]
	End    reader.Position
}

func (self *FuncDecl) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}
