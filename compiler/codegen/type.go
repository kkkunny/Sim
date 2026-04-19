package codegen

import (
	"fmt"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildType(t hir.Type) cir.Type {
	switch t := t.(type) {
	case *hir.UnitType:
		return cir.Void
	case *hir.SintType:
		switch t.Bits {
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
	case *hir.UintType:
		switch t.Bits {
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
	case *hir.FloatType:
		switch t.Bits {
		case 32:
			return cir.F32
		case 64:
			return cir.F64
		default:
			panic("unreachable")
		}
	case *hir.FuncType:
		return c.buildFuncType(t)
	case *hir.TupleType:
		return c.buildTupleType(t)
	default:
		panic("unreachable")
	}
}

// 原生函数类型
func (c *CodeGenerator) buildNativeFuncType(t *hir.FuncType) *cir.FuncType {
	r := c.buildType(t.Return)
	ps := stlslices.Map(t.Params, func(i int, e hir.Type) cir.Type {
		return c.buildType(e)
	})
	return &cir.FuncType{
		Return: r,
		Params: ps,
	}
}

// 函数指针类型
func (c *CodeGenerator) buildFuncPointerType(t *hir.FuncType) *cir.AliasType {
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
func (c *CodeGenerator) buildFuncType(t *hir.FuncType) *cir.MacroType {
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

func (c *CodeGenerator) buildTupleType(t *hir.TupleType) *cir.StructType {
	fields := make([]*cir.StructTypeField, len(t.Elems))
	for i, e := range t.Elems {
		fn := fmt.Sprintf("_f%d", i)
		ft := c.buildType(e)
		fields[i] = cir.NewStructTypeField(ft, fn)
	}
	return cir.NewStructType("", fields...)
}
