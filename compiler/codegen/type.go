package codegen

import "github.com/kkkunny/Sim/compiler/ast"

func (c *CodeGenerator) genType(t ast.Type) string {
	switch t := t.(type) {
	case *ast.IdentType:
		switch t.Name {
		case "i32":
			return "int"
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}
