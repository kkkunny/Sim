package analyze

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (a *Analyzer) analyzeType(t ast.Type) hir.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		switch t.Name.OriginText {
		case "i8":
			return hir.I8
		case "i16":
			return hir.I16
		case "i32":
			return hir.I32
		case "i64":
			return hir.I64
		case "unit":
			return hir.Unit
		default:
			panic(fmt.Sprintf("unknown type: %s", t.Name.OriginText))
		}
	default:
		panic("unsupported type")
	}
}
