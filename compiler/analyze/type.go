package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeType(t ast.Type) hir.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		switch t.Name.OriginText {
		case "unit":
			return hir.Unit
		case "i8":
			return hir.I8
		case "i16":
			return hir.I16
		case "i32":
			return hir.I32
		case "i64":
			return hir.I64
		case "u8":
			return hir.U8
		case "u16":
			return hir.U16
		case "u32":
			return hir.U32
		case "u64":
			return hir.U64
		case "f32":
			return hir.F32
		case "f64":
			return hir.F64
		default:
			a.reporter.Fatalf(
				t.Name.Position,
				report.Errors.UnknownIdentifier,
				t.Name.OriginText,
			)
			return nil
		}
	default:
		panic("unreachable")
	}
}
