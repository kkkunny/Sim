package ast

import (
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/token"
)

type Attribute interface {
	printWriter
	AttrName() string
	Position() reader.Position
}

type Extern struct {
	BeginPosition reader.Position
	Name          token.Token
	EndPosition   reader.Position
}

func (a *Extern) AttrName() string {
	return "extern"
}

func (a *Extern) print(p *printer) {
	p.WriteString("@extern(")
	p.WriteToken(a.Name)
	p.WriteString(")")
}

func (a *Extern) Position() reader.Position {
	return reader.MixPosition(a.BeginPosition, a.EndPosition)
}
