package ast

import (
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
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

type GenericNameDef struct {
	Name token.Token
	ParamBegin reader.Position
	Params []token.Token
	ParamEnd reader.Position
}

func (self GenericNameDef) Position() reader.Position{
	return reader.MixPosition(self.Name.Position, self.ParamEnd)
}
