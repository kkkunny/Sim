package ast

import "github.com/kkkunny/Sim/reader"

// Ast 抽象语法树
type Ast interface {
	Position() reader.Position
}
