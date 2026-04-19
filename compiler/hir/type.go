package hir

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
)

type Type interface {
	printWriter
	fmt.Stringer
	typ()
	Equal(Type) bool
}

var Unit = &UnitType{}

type UnitType struct{}

func (*UnitType) typ() {}

func (t *UnitType) print(p *printer) {
	p.WriteString(t.String())
}

func (t *UnitType) String() string {
	return "unit"
}

func (t *UnitType) Equal(p Type) bool {
	return stlval.Is[*UnitType](p)
}

type IntegerType interface {
	Type
	GetBits() uint8
	integer()
}

var (
	I8  = &SintType{Bits: 8}
	I16 = &SintType{Bits: 16}
	I32 = &SintType{Bits: 32}
	I64 = &SintType{Bits: 64}
)

type SintType struct {
	Bits uint8
}

func (*SintType) typ() {}

func (t *SintType) print(p *printer) {
	p.WriteFormat(t.String())
}

func (t *SintType) String() string {
	return fmt.Sprintf("i%d", t.Bits)
}

func (t *SintType) Equal(p Type) bool {
	dst, ok := p.(*SintType)
	if !ok {
		return false
	}
	return t.Bits == dst.Bits
}

func (t *SintType) GetBits() uint8 {
	return t.Bits
}

func (t *SintType) integer() {}

var (
	U8  = &UintType{Bits: 8}
	U16 = &UintType{Bits: 16}
	U32 = &UintType{Bits: 32}
	U64 = &UintType{Bits: 64}
)

type UintType struct {
	Bits uint8
}

func (*UintType) typ() {}

func (t *UintType) print(p *printer) {
	p.WriteFormat(t.String())
}

func (t *UintType) String() string {
	return fmt.Sprintf("u%d", t.Bits)
}

func (t *UintType) Equal(p Type) bool {
	dst, ok := p.(*UintType)
	if !ok {
		return false
	}
	return t.Bits == dst.Bits
}

func (t *UintType) GetBits() uint8 {
	return t.Bits
}

func (t *UintType) integer() {}

var (
	F32 = &FloatType{Bits: 32}
	F64 = &FloatType{Bits: 64}
)

type FloatType struct {
	Bits uint8
}

func (*FloatType) typ() {}

func (t *FloatType) print(p *printer) {
	p.WriteFormat(t.String())
}

func (t *FloatType) String() string {
	return fmt.Sprintf("f%d", t.Bits)
}

func (t *FloatType) Equal(p Type) bool {
	dst, ok := p.(*FloatType)
	if !ok {
		return false
	}
	return t.Bits == dst.Bits
}

func (t *FloatType) GetBits() uint8 {
	return t.Bits
}

type FuncType struct {
	Return Type
	Params []Type
}

func NewFuncType(ret Type, params ...Type) *FuncType {
	return &FuncType{
		Return: ret,
		Params: params,
	}
}

func (*FuncType) typ() {}

func (t *FuncType) print(p *printer) {
	p.WriteString("(")
	for i, param := range t.Params {
		p.WriteBy(param)
		if i < len(t.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	if t.Return != Unit {
		p.WriteString(" -> ")
		p.WriteBy(t.Return)
	}
}

func (t *FuncType) String() string {
	var buf strings.Builder
	t.print(newPrint(&buf))
	return buf.String()
}

func (t *FuncType) Equal(p Type) bool {
	dst, ok := p.(*FuncType)
	if !ok {
		return false
	}
	if !t.Return.Equal(dst.Return) || len(t.Params) != len(dst.Params) {
		return false
	}
	return stlslices.All(t.Params, func(i int, p Type) bool {
		return p.Equal(dst.Params[i])
	})
}

type TupleType struct {
	Elems []Type
}

func NewTupleType(elems ...Type) *TupleType {
	return &TupleType{
		Elems: elems,
	}
}

func (*TupleType) typ() {}

func (t *TupleType) print(p *printer) {
	p.WriteString("(")
	for i, param := range t.Elems {
		p.WriteBy(param)
		if i < len(t.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (t *TupleType) String() string {
	var buf strings.Builder
	t.print(newPrint(&buf))
	return buf.String()
}

func (t *TupleType) Equal(p Type) bool {
	dst, ok := p.(*TupleType)
	if !ok {
		return false
	}
	if len(t.Elems) != len(dst.Elems) {
		return false
	}
	return stlslices.All(t.Elems, func(i int, p Type) bool {
		return p.Equal(dst.Elems[i])
	})
}
