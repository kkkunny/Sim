package ast

import (
	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Attr 属性
type Attr interface {
	Ast
	AttrName() string
}

type Extern struct {
	Begin reader.Position
	Name  token.Token
	End   reader.Position
}

func (self *Extern) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Extern) AttrName() string {
	return "extern"
}

type Inline struct {
	Begin reader.Position
	End   reader.Position
}

func (self *Inline) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Inline) AttrName() string {
	return "inline"
}

type NoInline struct {
	Begin reader.Position
	End   reader.Position
}

func (self *NoInline) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *NoInline) AttrName() string {
	return "noinline"
}

type VarArg struct {
	Begin reader.Position
	End   reader.Position
}

func (self *VarArg) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *VarArg) AttrName() string {
	return "var_arg"
}
