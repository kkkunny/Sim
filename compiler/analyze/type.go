package analyze

import (
	"math/big"
	"strconv"

	stlslices "github.com/kkkunny/stl/container/slices"

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
	case *ast.FuncType:
		params := stlslices.Map(t.Params, func(_ int, p ast.Type) hir.Type {
			return a.analyzeType(p)
		})
		var rt hir.Type = hir.Unit
		if returnType, ok := t.ReturnType.Value(); ok {
			rt = a.analyzeType(returnType)
		}
		return hir.NewFuncType(rt, params...)
	case *ast.TupleType:
		elems := stlslices.Map(t.Elems, func(_ int, p ast.Type) hir.Type {
			return a.analyzeType(p)
		})
		return hir.NewTupleType(elems...)
	case *ast.ArrayType:
		v, _ := strconv.ParseInt(t.Size.OriginText, 10, 64)
		size := big.NewInt(v)
		elem := a.analyzeType(t.Elem)
		return hir.NewArrayType(size, elem)
	default:
		panic("unreachable")
	}
}
