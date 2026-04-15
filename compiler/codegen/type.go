package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildType(t hir.Type) cir.Type {
	switch t.(type) {
	case *hir.IntType:
		return cir.SInt
	case *hir.UnitType:
		return cir.Void
	default:
		panic("unreachable")
	}
}
