package ast

import (
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/token"
)

type Type interface {
	typ()
	printWriter
	Position() reader.Position
}

type IdentType struct {
	Name token.Token
}

func (t *IdentType) typ() {}

func (t *IdentType) print(p *printer) {
	p.WriteToken(t.Name)
}

func (t *IdentType) Position() reader.Position {
	return t.Name.Position
}
