package codegen

import (
	stlslices "github.com/kkkunny/stl/container/slices"

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
		key := t.String()
		at, ok := c.typeCache[key]
		if ok {
			return at
		}

		r := c.buildType(t.Return)
		ps := stlslices.Map(t.Params, func(i int, e hir.Type) cir.Type {
			return c.buildType(e)
		})
		ft := &cir.PointerType{
			Elem: &cir.FuncType{
				Return: r,
				Params: ps,
			},
		}

		at = &cir.AliasType{
			Define: c.builder.BuildTypedef(ft, ""),
		}
		c.typeCache[key] = at
		return at
	default:
		panic("unreachable")
	}
}
