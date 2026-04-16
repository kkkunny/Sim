package cir

import (
	"github.com/kkkunny/stl/enum"
)

type Type interface {
	printWriter
	printWithName(*printer, string)
	typ()
}

var Void = &VoidType{}

type VoidType struct{}

func (*VoidType) typ() {}

func (t *VoidType) print(p *printer) {
	p.WriteString("void")
}

func (t *VoidType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

var (
	I8  = &IntType{Kind: IntTypeKindEnum.I8}
	I16 = &IntType{Kind: IntTypeKindEnum.I16}
	I32 = &IntType{Kind: IntTypeKindEnum.I32}
	I64 = &IntType{Kind: IntTypeKindEnum.I64}
)

type IntTypeKind string

var IntTypeKindEnum = enum.New[struct {
	I8  IntTypeKind `enum:"i8"`
	I16 IntTypeKind `enum:"i16"`
	I32 IntTypeKind `enum:"i32"`
	I64 IntTypeKind `enum:"i64"`
}]()

type IntType struct {
	Kind IntTypeKind
}

func (*IntType) typ() {}

func (t *IntType) print(p *printer) {
	p.WriteString(string(t.Kind))
}

func (t *IntType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

var (
	F32 = &FloatType{Kind: FloatTypeKindEnum.F32}
	F64 = &FloatType{Kind: FloatTypeKindEnum.F64}
)

type FloatTypeKind string

var FloatTypeKindEnum = enum.New[struct {
	F32 FloatTypeKind `enum:"f32"`
	F64 FloatTypeKind `enum:"f64"`
}]()

type FloatType struct {
	Kind FloatTypeKind
}

func (*FloatType) typ() {}

func (t *FloatType) print(p *printer) {
	p.WriteString(string(t.Kind))
}

func (t *FloatType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

type FuncType struct {
	Return Type
	Params []Type
}

func (*FuncType) typ() {}

func (t *FuncType) print(*printer) {}

func (t *FuncType) printWithName(p *printer, name string) {
	p.WriteBy(t.Return)
	p.WriteString("(")
	p.WriteFormat(name)
	p.WriteString(")")
}
