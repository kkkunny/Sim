package c

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) codegenType(ir mir.Type) string {
	switch t := ir.(type) {
	case mir.VoidType:
		return self.codegenVoidType()
	case mir.SintType:
		return self.codegenSintType(t)
	case mir.UintType:
		return self.codegenUintType(t)
	case mir.FloatType:
		return self.codegenFloatType(t)
	case mir.PtrType:
		return self.codegenPtrType(t)
	case mir.FuncType:
		return self.codegenFuncTypePtr(t)
	case mir.ArrayType:
		return self.codegenArrayType(t)
	case mir.StructType:
		return self.codegenStructType(t)
	default:
		panic("unreachable")
	}
}

func (self *COutputer) codegenVoidType() string {
	return "void"
}

func (self *COutputer) codegenSintType(ir mir.SintType) string {
	key := fmt.Sprintf("i%d", ir.Bits())
	if self.typedefs.ContainKey(key) {
		return self.typedefs.Get(key).First
	}
	var ctype string
	switch stlmath.RoundTo(ir.Size(), stlos.Size(ir.Align())*8) {
	case 8:
		ctype = "signed char"
	case 16:
		ctype = "signed short"
	case 32:
		ctype = "signed int"
	case 64:
		ctype = "signed long long"
	default:
		panic("unreachable")
	}
	name := fmt.Sprintf("_sim_type_i%d", ir.Bits())
	self.typedefs.Set(key, pair.NewPair(name, fmt.Sprintf("typedef %s %s", ctype, name)))
	return name
}

func (self *COutputer) codegenUintType(ir mir.UintType) string {
	key := fmt.Sprintf("u%d", ir.Bits())
	if self.typedefs.ContainKey(key) {
		return self.typedefs.Get(key).First
	}
	var ctype string
	switch stlmath.RoundTo(ir.Size(), stlos.Size(ir.Align())*8) {
	case 8:
		ctype = "unsigned char"
	case 16:
		ctype = "unsigned short"
	case 32:
		ctype = "unsigned int"
	case 64:
		ctype = "unsigned long long"
	default:
		panic("unreachable")
	}
	name := fmt.Sprintf("_sim_type_u%d", ir.Bits())
	self.typedefs.Set(key, pair.NewPair(name, fmt.Sprintf("typedef %s %s", ctype, name)))
	return name
}

func (self *COutputer) codegenFloatType(ir mir.FloatType) string {
	key := fmt.Sprintf("f%d", ir.Bits())
	if self.typedefs.ContainKey(key) {
		return self.typedefs.Get(key).First
	}
	var ctype string
	switch stlmath.RoundTo(ir.Size(), stlos.Size(ir.Align())*8) {
	case 32:
		ctype = "float"
	case 64:
		ctype = "double"
	case 128:
		ctype = "long double"
	default:
		panic("unreachable")
	}
	name := fmt.Sprintf("_sim_type_f%d", ir.Bits())
	self.typedefs.Set(key, pair.NewPair(name, fmt.Sprintf("typedef %s %s", ctype, name)))
	return name
}

func (self *COutputer) codegenPtrType(ir mir.PtrType) string {
	return self.codegenType(ir.Elem()) + "*"
}

func (self *COutputer) codegenFuncTypePtr(ir mir.FuncType) string {
	ret := self.codegenType(ir.Ret())
	params := stlslices.Map(ir.Params(), func(_ int, e mir.Type) string {
		return self.codegenType(e)
	})
	key := fmt.Sprintf("%s(%s)", ret, strings.Join(params, ", "))
	if self.typedefs.ContainKey(key) {
		return self.typedefs.Get(key).First
	}
	name := fmt.Sprintf("_sim_type_%d", self.typedefs.Length())
	self.typedefs.Set(key, pair.NewPair(name, fmt.Sprintf("%s(*%s)(%s)", ret, name, strings.Join(params, ", "))))
	return name
}

func (self *COutputer) codegenArrayType(ir mir.ArrayType) string {
	elem := self.codegenType(ir.Elem())
	key := fmt.Sprintf("%s[%d]", elem, ir.Size())
	if self.typedefs.ContainKey(key) {
		return self.typedefs.Get(key).First
	}
	name := fmt.Sprintf("_sim_type_%d", self.typedefs.Length())
	self.typedefs.Set(key, pair.NewPair(name, fmt.Sprintf("typedef struct{%s data[%d];}%s", elem, ir.Size(), name)))
	return name
}

func (self *COutputer) codegenStructType(ir mir.StructType) string {
	switch st := ir.(type) {
	case *mir.UnnamedStructType:
		var buf strings.Builder
		buf.WriteString("struct{")
		for i, e := range ir.Elems() {
			buf.WriteString(self.codegenType(e))
			buf.WriteByte(' ')
			buf.WriteString(fmt.Sprintf("e%d", i))
			buf.WriteByte(';')
		}
		buf.WriteByte('}')
		return buf.String()
	case *mir.NamedStruct:
		return "struct " + self.types.Get(st)
	default:
		panic("unreachable")
	}
}
