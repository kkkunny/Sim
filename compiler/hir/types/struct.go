package types

import (
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type StructField struct {
	Pub  bool
	Mut  bool
	Name string
	Type Type
}

type StructType interface {
	Type
	structType()
	GetFields() []*StructField
}

type _StructType struct {
	Fields []*StructField
}

func NewStructType(fields ...*StructField) StructType {
	return &_StructType{Fields: fields}
}

func (t *_StructType) structType() {}

func (t *_StructType) Print(p *hir.Printer) {
	p.WriteString("struct {")
	if len(t.Fields) > 0 {
		p.NextLine(+1)
		for i, f := range t.Fields {
			if f.Pub {
				p.WriteString("pub ")
			}
			if f.Mut {
				p.WriteString("mut ")
			}
			p.WriteString(f.Name)
			p.WriteString(": ")
			p.WriteBy(f.Type)
			if i < len(t.Fields)-1 {
				p.NextLine()
			}
		}
		p.NextLine(-1)
	}
	p.WriteString("}")
}

func (t *_StructType) String() string {
	var buf strings.Builder
	hir.Print(&buf, t)
	return buf.String()
}

func (t *_StructType) Equal(p Type) bool {
	st, ok := p.(StructType)
	if !ok {
		return false
	}
	fields := st.GetFields()
	if len(t.Fields) != len(fields) {
		return false
	}
	return stlslices.All(t.Fields, func(i int, f *StructField) bool {
		return f.Name == fields[i].Name && f.Type.Equal(fields[i].Type)
	})
}

func (t *_StructType) GetFields() []*StructField {
	return t.Fields
}

type _CustomStructType struct {
	_CustomBaseType[StructType]
}

func (t *_CustomStructType) structType() {}

func (t *_CustomStructType) GetFields() []*StructField {
	return t.Underlying.GetFields()
}
