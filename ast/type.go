package ast

import (
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
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
