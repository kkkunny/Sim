package codegen

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildType(t hir.Type) cir.Type {
	switch t := t.(type) {
	case *hir.IntType:
		return cir.SInt
	case *hir.UnitType:
		return cir.Void
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
