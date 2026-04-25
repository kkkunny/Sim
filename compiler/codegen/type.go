package codegen

import (
	"fmt"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) buildType(t types.Type) cir.Type {
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
		return c.buildFuncType(t)
	case types.TupleType:
		return c.buildTupleType(t)
	case types.ArrayType:
		return c.buildArrayType(t)
	case types.UnionType:
		return c.buildUnionType(t)
	case types.RefType:
		return c.buildRefType(t)
	default:
		panic("unreachable")
	}
}

// 原生函数类型
func (c *CodeGenerator) buildNativeFuncType(t types.FuncType) *cir.FuncType {
	r := c.buildType(t.GetReturn())
	ps := stlslices.Map(t.GetParams(), func(i int, e types.Type) cir.Type {
		return c.buildType(e)
	})
	return &cir.FuncType{
		Return: r,
		Params: ps,
	}
}

// 函数指针类型
func (c *CodeGenerator) buildFuncPointerType(t types.FuncType) *cir.AliasType {
	key := t.String()
	at, ok := c.typeCache[key]
	if ok {
		return at
	}

	ft := c.buildNativeFuncType(t)
	at = cir.NewAliasType(c.builder.BuildTypedef(cir.NewPointerType(ft), ""))
	c.typeCache[key] = at
	return at
}

// 函数胖类型，用于变量定义、赋值
func (c *CodeGenerator) buildFuncType(t types.FuncType) *cir.MacroType {
	ft := c.buildFuncPointerType(t)

	key := fmt.Sprintf("closure:%s", t)
	ct, ok := c.typeCache[key]
	if !ok {
		nativeFt := stlval.Ptr(*ft.Type.(*cir.PointerType).Elem.(*cir.FuncType))
		nativeFt.Params = append([]cir.Type{cir.VoidPtr}, nativeFt.Params...)
		ct = cir.NewAliasType(c.builder.BuildTypedef(cir.NewPointerType(nativeFt), ""))
		c.typeCache[key] = ct
	}

	return cir.NewMacroType("FUNCTYPE", ft, ct)
}

func (c *CodeGenerator) buildTupleType(t types.TupleType) *cir.StructType {
	fields := make([]*cir.Member, len(t.GetElems()))
	for i, e := range t.GetElems() {
		fn := fmt.Sprintf("_f%d", i+1)
		ft := c.buildType(e)
		fields[i] = cir.NewMember(ft, fn)
	}
	return cir.NewStructType("", fields...)
}

func (c *CodeGenerator) buildArrayType(t types.ArrayType) *cir.AliasType {
	key := t.String()
	at, ok := c.typeCache[key]
	if ok {
		return at
	}

	elem := c.buildType(t.GetElem())
	at = cir.NewAliasType(c.builder.BuildTypedef(cir.NewStructType("", cir.NewMember(cir.NewArrayType(elem, t.GetSize()), "array")), ""))
	c.typeCache[key] = at
	return at
}

func (c *CodeGenerator) buildUnionType(t types.UnionType) *cir.AliasType {
	key := t.String()
	ut, ok := c.typeCache[key]
	if ok {
		return ut
	}

	elems := stlslices.Map(t.GetElems(), func(_ int, e types.Type) cir.Type {
		return c.buildType(e)
	})
	ut = cir.NewAliasType(c.builder.BuildTypedef(cir.NewStructType(
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
	), ""))
	c.typeCache[key] = ut
	return ut
}

func (c *CodeGenerator) buildRefType(t types.RefType) *cir.PointerType {
	elem := c.buildType(t.PtrTo())
	return cir.NewPointerType(elem)
}
