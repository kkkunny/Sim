package cir

import (
	"fmt"
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
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

func NewFuncType(ret Type, param ...Type) *FuncType {
	return &FuncType{
		Return: ret,
		Params: param,
	}
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
	Def *Typedef
}

func NewAliasType(def *Typedef) *AliasType {
	return &AliasType{Def: def}
}

func (t *AliasType) stmt() {}

func (t *AliasType) zeroSize() bool {
	return t.Def.Type.zeroSize()
}

func (t *AliasType) print(p *printer) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString(t.Def.Name)
}

func (t *AliasType) printWithName(p *printer, name string) {
	if t.zeroSize() {
		p.WriteString("ZERO_TYPE")
		return
	}

	p.WriteString(t.Def.Name)
	p.WriteString(" ")
	p.WriteFormat(name)
}

var (
	I8  = NewMacroType("i8")
	I16 = NewMacroType("i16")
	I32 = NewMacroType("i32")
	I64 = NewMacroType("i64")
	U8  = NewMacroType("u8")
	U16 = NewMacroType("u16")
	U32 = NewMacroType("u32")
	U64 = NewMacroType("u64")

	F32 = NewMacroType("f32")
	F64 = NewMacroType("f64")

	Bool = NewMacroType("bool")
)

type MacroType struct {
	Name string
	Args []any
}

func NewMacroType(name string, args ...any) *MacroType {
	return &MacroType{Name: name, Args: args}
}

func (t *MacroType) zeroSize() bool {
	return false
}

func (t *MacroType) print(p *printer) {
	p.WriteString(t.Name)
	if len(t.Args) > 0 {
		p.WriteString("(")
		for i, arg := range t.Args {
			if arg != nil {
				if pp, ok := arg.(printWriter); ok {
					p.WriteBy(pp)
				} else {
					p.WriteString(fmt.Sprintf("%v", arg))
				}
			}
			if i < len(t.Args)-1 {
				p.WriteString(", ")
			}
		}
		p.WriteString(")")
	}
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
	Fields optional.Optional[[]*Member]
}

func NewStructType(name string, fields optional.Optional[[]*Member]) *StructType {
	return &StructType{Name: name, Fields: fields}
}

func (t *StructType) zeroSize() bool {
	if t.Fields.IsNone() {
		return false
	} else if len(t.Fields.MustValue()) == 0 {
		return true
	}
	return stlslices.All(t.Fields.MustValue(), func(_ int, m *Member) bool {
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
	if fields, ok := t.Fields.Value(); ok {
		p.WriteString("{")
		if len(fields) > 0 {
			p.NextLine(+1)
			for i, f := range fields {
				f.Type.printWithName(p, f.Name)
				p.WriteString(";")
				if i < len(fields)-1 {
					p.NextLine()
				} else {
					p.NextLine(-1)
				}
			}
		}
		p.WriteString("}")
	}
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
