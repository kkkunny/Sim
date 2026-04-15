package analyse

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (a *Analyzer) analyseType(t ast.Type) hir.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		switch t.Name.OriginText {
		case "i32":
			return hir.I32
		case "unit":
			return hir.Unit
		default:
			panic(fmt.Sprintf("unknown type: %s", t.Name.OriginText))
		}
	default:
		panic("unsupported type")
	}
}
