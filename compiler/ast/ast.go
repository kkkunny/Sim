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
