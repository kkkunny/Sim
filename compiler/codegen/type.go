package codegen

import (
	"fmt"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genType(t hir.Type) cir.Type {
	switch t := t.(type) {
	case types.CustomType:
		if t, ok := c.ctx.typeCache[t.GetName()]; ok {
			return t
		}
		return c.genCustomTypeDef(t.GetDef())
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
	case types.StringType:
		return cir.NewMacroType("str")
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
	case types.StructType:
		return c.genStructType(t)
	default:
		panic("unreachable")
	}
}

// 原生函数类型
func (c *CodeGenerator) genNativeFuncType(t types.FuncType) *cir.FuncType {
	r := c.genType(t.GetReturn())
	ps := stlslices.Map(t.GetParams(), func(i int, e hir.Type) cir.Type {
		return c.genType(e)
	})
	return cir.NewFuncType(r, ps...)
}

// 函数胖类型，用于变量定义、赋值
func (c *CodeGenerator) genFuncType(t types.FuncType) *cir.AliasType {
	key := t.String()
	at, ok := c.ctx.typeCache[key]
	if ok {
		return at
	}

	ft := c.genNativeFuncType(t)

	at = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewMacroType("FUNC_TYPE", append([]any{ft.Return}, stlslices.AsAny(ft.Params)...)...), "")))
	c.ctx.typeCache[key] = at
	return at
}

func (c *CodeGenerator) genTupleType(t types.TupleType) *cir.AliasType {
	key := t.String()
	at, ok := c.ctx.typeCache[key]
	if ok {
		return at
	}

	at = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(c.genFlatTupleType(t), "")))
	c.ctx.typeCache[key] = at
	return at
}

func (c *CodeGenerator) genFlatTupleType(t types.TupleType) *cir.StructType {
	fields := make([]*cir.Member, len(t.GetElems()))
	for i, e := range t.GetElems() {
		fn := fmt.Sprintf("e%d", i+1)
		ft := c.genType(e)
		fields[i] = cir.NewMember(ft, fn)
	}
	return cir.NewStructType("", optional.Some(fields))
}

func (c *CodeGenerator) genArrayType(t types.ArrayType) *cir.AliasType {
	key := t.String()
	at, ok := c.ctx.typeCache[key]
	if ok {
		return at
	}

	at = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(c.genFlatArrayType(t), "")))
	c.ctx.typeCache[key] = at
	return at
}

func (c *CodeGenerator) genFlatArrayType(t types.ArrayType) *cir.MacroType {
	elem := c.genType(t.GetElem())
	return cir.NewMacroType("ARRAY_TYPE", elem, cir.NewInteger(t.GetSize()))
}

func (c *CodeGenerator) genUnionType(t types.UnionType) *cir.AliasType {
	key := t.String()
	ut, ok := c.ctx.typeCache[key]
	if ok {
		return ut
	}

	ut = cir.NewAliasType(cir.BuildStmt(c.builder, cir.NewTypedef(c.genFlatUnionType(t), "")))
	c.ctx.typeCache[key] = ut
	return ut
}

func (c *CodeGenerator) genFlatUnionType(t types.UnionType) *cir.StructType {
	elems := stlslices.Map(t.GetElems(), func(_ int, e hir.Type) cir.Type {
		return c.genType(e)
	})
	return cir.NewStructType(
		"",
		optional.Some([]*cir.Member{
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
		}),
	)
}

func (c *CodeGenerator) genRefType(t types.RefType) *cir.MacroType {
	elem := c.genType(t.PtrTo())
	return cir.NewMacroType("PTR_TYPE", elem)
}

func (c *CodeGenerator) genStructType(t types.StructType) *cir.StructType {
	fields := make([]*cir.Member, len(t.GetFields()))
	for i, f := range t.GetFields() {
		fields[i] = cir.NewMember(c.genType(f.Type), f.Name)
	}
	return cir.NewStructType("", optional.Some(fields))
}
