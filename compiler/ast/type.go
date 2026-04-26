package ast

import (
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

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

// TupleType 元组类型，Elems数量不可能为1
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

type ArrayType struct {
	BeginPosition reader.Position
	Size          token.Token
	Elem          Type
}

func (t *ArrayType) typ() {}

func (t *ArrayType) print(p *printer) {
	p.WriteString("[")
	p.WriteString(t.Size.OriginText)
	p.WriteString("]")
	p.WriteBy(t.Elem)
}

func (t *ArrayType) Position() reader.Position {
	return reader.MixPosition(t.BeginPosition, t.Elem.Position())
}

// UnionType 联合类型，Elems数量不可能小于2
type UnionType struct {
	Elems []Type
}

func (t *UnionType) typ() {}

func (t *UnionType) print(p *printer) {
	for i, elem := range t.Elems {
		switch elem := elem.(type) {
		case *FuncType:
			if i < len(t.Elems)-1 {
				p.WriteString("(")
				p.WriteBy(elem)
				p.WriteString(")")
			} else {
				p.WriteBy(elem)
			}
		case *UnionType:
			p.WriteString("(")
			p.WriteBy(elem)
			p.WriteString(")")
		default:
			p.WriteBy(elem)
		}
		if i < len(t.Elems)-1 {
			p.WriteString(" | ")
		}
	}
}

func (t *UnionType) Position() reader.Position {
	return reader.MixPosition(stlslices.First(t.Elems).Position(), stlslices.Last(t.Elems).Position())
}

type RefType struct {
	BeginPosition reader.Position
	Mut           bool
	Elem          Type
}

func (t *RefType) typ() {}

func (t *RefType) print(p *printer) {
	p.WriteString("&")
	if t.Mut {
		p.WriteString("mut ")
	}
	p.WriteBy(t.Elem)
}

func (t *RefType) Position() reader.Position {
	return reader.MixPosition(t.BeginPosition, t.Elem.Position())
}
