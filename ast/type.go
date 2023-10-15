package ast

import (
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
type IdentType struct {
	Name token.Token
}

func (self *IdentType) Position() reader.Position {
	return self.Name.Position
}

func (self *IdentType) typ() {}

type FuncType struct {
	Begin reader.Position
	Ret   util.Option[Type]
	End   reader.Position
}

func (self *FuncType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *FuncType) typ() {}
