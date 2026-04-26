package types

import (
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type UnionType interface {
	Type
	GetElems() []Type
	union()
}

type _UnionType struct {
	Elems []Type
}

func NewUnionType(elems ...Type) UnionType {
	return &_UnionType{Elems: elems}
}

func (*_UnionType) union() {}

func (t *_UnionType) Print(p *hir.Printer) {
	for i, elem := range t.Elems {
		switch elem := elem.(type) {
		case FuncType:
			if i < len(t.Elems)-1 {
				p.WriteString("(")
				p.WriteBy(elem)
				p.WriteString(")")
			} else {
				p.WriteBy(elem)
			}
		case UnionType:
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

func (t *_UnionType) String() string {
	var buf strings.Builder
	hir.Print(&buf, t)
	return buf.String()
}

func (t *_UnionType) Equal(p Type) bool {
	dst, ok := p.(UnionType)
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

func (t *_UnionType) GetElems() []Type {
	return t.Elems
}

type _CustomUnionType struct {
	_CustomBaseType[UnionType]
}

func (*_CustomUnionType) union() {}

func (t *_CustomUnionType) GetElems() []Type {
	return t.Underlying.GetElems()
}
