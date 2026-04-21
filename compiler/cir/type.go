package cir

import (
	"math/big"

	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"
)

type Type interface {
	printWriter
	printWithName(*printer, string)
	zeroSize() bool
}

var Void = &VoidType{}

type VoidType struct{}

func (*VoidType) zeroSize() bool {
	panic("unreachable")
}

func (t *VoidType) print(p *printer) {
	p.WriteString("void")
}

func (t *VoidType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

var (
	I8  = &IntegerType{Kind: IntTypeKindEnum.I8}
	I16 = &IntegerType{Kind: IntTypeKindEnum.I16}
	I32 = &IntegerType{Kind: IntTypeKindEnum.I32}
	I64 = &IntegerType{Kind: IntTypeKindEnum.I64}
	U8  = &IntegerType{Kind: IntTypeKindEnum.U8}
	U16 = &IntegerType{Kind: IntTypeKindEnum.U16}
	U32 = &IntegerType{Kind: IntTypeKindEnum.U32}
	U64 = &IntegerType{Kind: IntTypeKindEnum.U64}
)

type IntTypeKind string

var IntTypeKindEnum = enum.New[struct {
	I8  IntTypeKind `enum:"i8"`
	I16 IntTypeKind `enum:"i16"`
	I32 IntTypeKind `enum:"i32"`
	I64 IntTypeKind `enum:"i64"`
	U8  IntTypeKind `enum:"u8"`
	U16 IntTypeKind `enum:"u16"`
	U32 IntTypeKind `enum:"u32"`
	U64 IntTypeKind `enum:"u64"`
}]()

type IntegerType struct {
	Kind IntTypeKind
}

func (*IntegerType) zeroSize() bool {
	return false
}

func (t *IntegerType) print(p *printer) {
	p.WriteString(string(t.Kind))
}

func (t *IntegerType) printWithName(p *printer, name string) {
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

func (*FloatType) zeroSize() bool {
	return false
}

func (t *FloatType) print(p *printer) {
	p.WriteString(string(t.Kind))
}

func (t *FloatType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

var VoidPtr = NewPointerType(Void)

type PointerType struct {
	Elem Type
}

func NewPointerType(elem Type) *PointerType {
	return &PointerType{Elem: elem}
}

func (*PointerType) zeroSize() bool {
	return false
}

func (t *PointerType) print(p *printer) {
	p.WriteBy(t.Elem)
	p.WriteString("*")
}

func (t *PointerType) printWithName(p *printer, name string) {
	t.Elem.printWithName(p, "*"+name)
}

type FuncType struct {
	Return Type
	Params []Type
}

func (*FuncType) zeroSize() bool {
	return false
}

func (t *FuncType) print(*printer) {
	panic("unreachable")
}

func (t *FuncType) printWithName(p *printer, name string) {
	p.WriteBy(t.Return)
	p.WriteString("(")
	p.WriteFormat(name)
	p.WriteString(")")
	p.WriteString("(")
	for i, param := range t.Params {
		p.WriteBy(param)
		if i < len(t.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

type AliasType struct {
	*Typedef
}

func NewAliasType(def *Typedef) *AliasType {
	return &AliasType{Typedef: def}
}

func (t *AliasType) zeroSize() bool {
	return t.Type.zeroSize()
}

func (t *AliasType) print(p *printer) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString(t.Name)
}

func (t *AliasType) printWithName(p *printer, name string) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString(t.Name)
	p.WriteString(" ")
	p.WriteFormat(name)
}

type MacroType struct {
	Name string
	Args []Type
}

func NewMacroType(name string, args ...Type) *MacroType {
	return &MacroType{Name: name, Args: args}
}

func (t *MacroType) zeroSize() bool {
	return false
}

func (t *MacroType) print(p *printer) {
	p.WriteString(t.Name)
	p.WriteString("(")
	for i, arg := range t.Args {
		p.WriteBy(arg)
		if i < len(t.Args)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (t *MacroType) printWithName(p *printer, name string) {
	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

type Member struct {
	Type Type
	Name string
}

func NewMember(t Type, name string) *Member {
	return &Member{
		Type: t,
		Name: name,
	}
}

type StructType struct {
	Name   string
	Fields []*Member
}

func NewStructType(name string, fields ...*Member) *StructType {
	return &StructType{Name: name, Fields: fields}
}

func (t *StructType) zeroSize() bool {
	if len(t.Fields) == 0 {
		return true
	}
	return stlslices.All(t.Fields, func(_ int, m *Member) bool {
		return m.Type.zeroSize()
	})
}

func (t *StructType) print(p *printer) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString("struct ")
	p.WriteString(t.Name)
	p.WriteString("{")
	if len(t.Fields) > 0 {
		p.NextLine(+1)
		for i, f := range t.Fields {
			f.Type.printWithName(p, f.Name)
			p.WriteString(";")
			if i < len(t.Fields)-1 {
				p.NextLine()
			} else {
				p.NextLine(-1)
			}
		}
	}
	p.WriteString("}")
}

func (t *StructType) printWithName(p *printer, name string) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}

type ArrayType struct {
	Elem Type
	Size *big.Int
}

func NewArrayType(elem Type, size *big.Int) *ArrayType {
	return &ArrayType{Elem: elem, Size: size}
}

func (t *ArrayType) zeroSize() bool {
	if t.Size.String() == "0" {
		return true
	}
	return t.Elem.zeroSize()
}

func (t *ArrayType) print(p *printer) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteBy(t.Elem)
	p.WriteString("[")
	p.WriteString(t.Size.String())
	p.WriteString("]")
}

func (t *ArrayType) printWithName(p *printer, name string) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteBy(t.Elem)
	p.WriteString(" ")
	p.WriteString(name)
	p.WriteString("[")
	p.WriteString(t.Size.String())
	p.WriteString("]")
}

type UnionType struct {
	Name    string
	Members []*Member
}

func NewUnionType(name string, members ...*Member) *UnionType {
	return &UnionType{Name: name, Members: members}
}

func (t *UnionType) zeroSize() bool {
	if len(t.Members) == 0 {
		return true
	}
	return stlslices.All(t.Members, func(i int, e *Member) bool {
		return e.Type.zeroSize()
	})
}

func (t *UnionType) print(p *printer) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString("union ")
	p.WriteString(t.Name)
	p.WriteString("{")
	if len(t.Members) > 0 {
		p.NextLine(+1)
		for i, m := range t.Members {
			m.Type.printWithName(p, m.Name)
			p.WriteString(";")
			if i < len(t.Members)-1 {
				p.NextLine()
			} else {
				p.NextLine(-1)
			}
		}
	}
	p.WriteString("}")
}

func (t *UnionType) printWithName(p *printer, name string) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteBy(t)
	p.WriteString(" ")
	p.WriteString(name)
}
