package types

import (
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type UnionType interface {
	hir.Type
	GetElems() []hir.Type
	union()
}

type _UnionType struct {
	Elems []hir.Type
}

func NewUnionType(elems ...hir.Type) UnionType {
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

func (t *_UnionType) Equal(p hir.Type) bool {
	dst, ok := p.(UnionType)
	if !ok {
		return false
	}
	dstElems := dst.GetElems()
	if len(t.Elems) != len(dstElems) {
		return false
	}
	return stlslices.All(t.Elems, func(i int, p hir.Type) bool {
		return p.Equal(dstElems[i])
	})
}

func (t *_UnionType) GetElems() []hir.Type {
	return t.Elems
}

type _CustomUnionType struct {
	_CustomBaseType[UnionType]
}

func (*_CustomUnionType) union() {}

func (t *_CustomUnionType) GetElems() []hir.Type {
	return t.GetUnderlying().(UnionType).GetElems()
}
