package ast

import (
	"github.com/kkkunny/stl/container/optional"

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

type FuncType struct {
	BeginPosition reader.Position
	Params        []Type
	ReturnType    optional.Optional[Type]
	EndPosition   reader.Position
}

func (t *FuncType) typ() {}

func (t *FuncType) print(p *printer) {
	p.WriteString("(")
	for i, param := range t.Params {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(param)
	}
	p.WriteString(")")
	if rt, ok := t.ReturnType.Value(); ok {
		p.WriteString(" -> ")
		p.WriteBy(rt)
	}
}

func (t *FuncType) Position() reader.Position {
	return reader.MixPosition(t.BeginPosition, t.EndPosition)
}

type TupleType struct {
	BeginPosition reader.Position
	Elems         []Type
	EndPosition   reader.Position
}

func (t *TupleType) typ() {}

func (t *TupleType) print(p *printer) {
	p.WriteString("(")
	for i, elem := range t.Elems {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(elem)
	}
	p.WriteString(")")
}

func (t *TupleType) Position() reader.Position {
	return reader.MixPosition(t.BeginPosition, t.EndPosition)
}
