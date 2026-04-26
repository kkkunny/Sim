package types

import (
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type TupleType interface {
	Type
	GetElems() []Type
	tuple()
}

type _TupleType struct {
	Elems []Type
}

func NewTupleType(elems ...Type) TupleType {
	return &_TupleType{
		Elems: elems,
	}
}

func (*_TupleType) tuple() {}

func (t *_TupleType) Print(p *hir.Printer) {
	p.WriteString("(")
	for i, param := range t.Elems {
		p.WriteBy(param)
		if i < len(t.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (t *_TupleType) String() string {
	var buf strings.Builder
	hir.Print(&buf, t)
	return buf.String()
}

func (t *_TupleType) Equal(p Type) bool {
	dst, ok := p.(TupleType)
	if !ok {
		return false
	}
	dstElems := dst.GetElems()
	if len(t.Elems) != len(dstElems) {
		return false
	}
	return stlslices.All(t.Elems, func(i int, p Type) bool {
		return p.Equal(dstElems[i])
	})
}

func (t *_TupleType) GetElems() []Type {
	return t.Elems
}

type _CustomTupleType struct {
	_CustomBaseType[TupleType]
}

func (*_CustomTupleType) tuple() {}

func (t *_CustomTupleType) GetElems() []Type {
	return t.Underlying.GetElems()
}
