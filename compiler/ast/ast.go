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
	Public bool
	Mutable bool
	Name    token.Token
	Type    Type
}

type List[T any] struct {
	Begin reader.Position
	Data  []T
	End   reader.Position
}

func (self List[T]) Position() reader.Position{
	return reader.MixPosition(self.Begin, self.End)
}

type GenericNameDef struct {
	Name token.Token
	Params util.Option[List[token.Token]]
}

func (self GenericNameDef) Position() reader.Position{
	if params, ok := self.Params.Value(); ok{
		return reader.MixPosition(self.Name.Position, params.Position())
	}
	return self.Name.Position
}

type GenericName struct {
	Name token.Token
	Params util.Option[List[Type]]
}

func (self GenericName) Position() reader.Position{
	if params, ok := self.Params.Value(); ok{
		return reader.MixPosition(self.Name.Position, params.Position())
	}
	return self.Name.Position
}
