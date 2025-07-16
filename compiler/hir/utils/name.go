package utils

import "github.com/kkkunny/Sim/compiler/token"

type Named interface {
	GetName() (Name, bool)
}

type NotGlobalNamed interface {
	NotGlobalNamed()
}

type Name struct {
	Position Position
	Value    string
}

func NewNameFromToken(t token.Token) Name {
	return Name{
		Value: t.Source(),
		Position: Position{
			Begin: t.Position.BeginOffset,
			End:   t.Position.EndOffset,
		},
	}
}

func (name Name) String() string {
	return name.Value
}
