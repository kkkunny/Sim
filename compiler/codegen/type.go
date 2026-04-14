package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
)

func (c *CodeGenerator) buildType(t ast.Type) cir.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		switch t.Name.OriginText {
		case "unit":
			return cir.Void
		case "i32":
			return cir.SInt
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}
