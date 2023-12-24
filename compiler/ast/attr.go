package ast

import (
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
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

type NoReturn struct {
	Begin reader.Position
	End   reader.Position
}

func (self *NoReturn) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *NoReturn) AttrName() string {
	return "noreturn"
}
