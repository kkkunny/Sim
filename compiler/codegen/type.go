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
	case *hir.IntType:
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
		r := c.buildType(t.Return)
		ps := stlslices.Map(t.Params, func(i int, e hir.Type) cir.Type {
			return c.buildType(e)
		})
		return &cir.FuncType{
			Return: r,
			Params: ps,
		}
	default:
		panic("unreachable")
	}
}
