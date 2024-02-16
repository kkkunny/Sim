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
	Mutable bool
	Name    token.Token
	Type    Type
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
