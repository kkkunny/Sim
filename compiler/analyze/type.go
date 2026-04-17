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

// 期待类型，两个类型必须完全相同
func (a *Analyzer) expectType(t ast.Type, expect hir.Type) hir.Type {
	get := a.analyzeType(t)
	if !get.Equal(expect) {
		a.reporter.Errorf(
			t.Position(),
			report.Errors.UnexpectedType,
			expect, get,
		)
	}
	return get
}
