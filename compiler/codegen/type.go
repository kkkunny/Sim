package codegen

import (
	"fmt"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genType(t types.Type) cir.Type {
	switch t := t.(type) {
	case types.UnitType:
		return cir.Void
	case types.SintType:
		switch t.GetBits() {
		case 8:
			return cir.I8
		case 16:
			return cir.I16
		case 32:
			return cir.I32
		case 64:
			return cir.I64
		default:
			panic("unreachable")
		}
	case types.UintType:
		switch t.GetBits() {
		case 8:
			return cir.U8
		case 16:
			return cir.U16
		case 32:
			return cir.U32
		case 64:
			return cir.U64
		default:
			panic("unreachable")
		}
	case types.FloatType:
		switch t.GetBits() {
		case 32:
			return cir.F32
		case 64:
			return cir.F64
		default:
			panic("unreachable")
		}
	case types.BooleanType:
		return cir.Bool
	case types.FuncType:
		return c.genFuncType(t)
	case types.TupleType:
		return c.genTupleType(t)
	case types.ArrayType:
		return c.genArrayType(t)
	case types.UnionType:
		return c.genUnionType(t)
	case types.RefType:
		return c.genRefType(t)
	default:
		panic("unreachable")
	}
}

// 原生函数类型
func (c *CodeGenerator) genNativeFuncType(t types.FuncType) *cir.FuncType {
	r := c.genType(t.GetReturn())
	ps := stlslices.Map(t.GetParams(), func(i int, e types.Type) cir.Type {
		return c.genType(e)
	})
	return &cir.FuncType{
		Return: r,
		Params: ps,
	}
}

// 函数指针类型
func (c *CodeGenerator) genFuncPointerType(t types.FuncType) *cir.AliasType {
	key := t.String()
	at, ok := c.typeCache[key]
	if ok {
		return at
	}

	ft := c.genNativeFuncType(t)
	at = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewPointerType(ft), "")))
	c.typeCache[key] = at
	return at
}

// 函数胖类型，用于变量定义、赋值
func (c *CodeGenerator) genFuncType(t types.FuncType) *cir.MacroType {
	ft := c.genFuncPointerType(t)

	key := fmt.Sprintf("closure:%s", t)
	ct, ok := c.typeCache[key]
	if !ok {
		nativeFt := stlval.Ptr(*ft.Def.Type.(*cir.PointerType).Elem.(*cir.FuncType))
		nativeFt.Params = append([]cir.Type{cir.VoidPtr}, nativeFt.Params...)
		ct = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewPointerType(nativeFt), "")))
		c.typeCache[key] = ct
	}

	return cir.NewMacroType("FUNCTYPE", ft, ct)
}

func (c *CodeGenerator) genTupleType(t types.TupleType) *cir.StructType {
	fields := make([]*cir.Member, len(t.GetElems()))
	for i, e := range t.GetElems() {
		fn := fmt.Sprintf("_f%d", i+1)
		ft := c.genType(e)
		fields[i] = cir.NewMember(ft, fn)
	}
	return cir.NewStructType("", fields...)
}

func (c *CodeGenerator) genArrayType(t types.ArrayType) *cir.AliasType {
	key := t.String()
	at, ok := c.typeCache[key]
	if ok {
		return at
	}

	elem := c.genType(t.GetElem())
	at = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewMacroType("ARRAY_TYPE", elem, cir.NewInteger(t.GetSize())), "")))
	c.typeCache[key] = at
	return at
}

func (c *CodeGenerator) genUnionType(t types.UnionType) *cir.AliasType {
	key := t.String()
	ut, ok := c.typeCache[key]
	if ok {
		return ut
	}

	elems := stlslices.Map(t.GetElems(), func(_ int, e types.Type) cir.Type {
		return c.genType(e)
	})
	ut = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewStructType(
		"",
		cir.NewMember(cir.U8, "t"),
		cir.NewMember(
			cir.NewUnionType(
				"",
				stlslices.Map(elems, func(i int, e cir.Type) *cir.Member {
					return cir.NewMember(e, fmt.Sprintf("t%d", i+1))
				})...,
			),
			"v",
		),
	), "")))
	c.typeCache[key] = ut
	return ut
}

func (c *CodeGenerator) genRefType(t types.RefType) *cir.PointerType {
	elem := c.genType(t.PtrTo())
	return cir.NewPointerType(elem)
}
