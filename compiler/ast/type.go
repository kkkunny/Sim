package ast

import (
	"github.com/kkkunny/Sim/compiler/token"
)

type Type interface {
	typ()
	printWriter
}

type IdentType struct {
	Name token.Token
}

func (t *IdentType) typ() {}

func (t *IdentType) print(p *printer) {
	p.WriteToken(t.Name)
}
