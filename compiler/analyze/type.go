package analyze

import (
	"math/big"
	"strconv"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeType(t ast.Type) types.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		if _, ok := a.scope.LookupType(t.Name.OriginText); ok {
			return a.analyzeTypeDef(t)
		}
		return a.analyzeBuildInIdentType(t)
	case *ast.FuncType:
		params := stlslices.Map(t.Params, func(_ int, p ast.Type) types.Type {
			return a.analyzeType(p)
		})
		var rt types.Type = types.Unit
		if returnType, ok := t.ReturnType.Value(); ok {
			rt = a.analyzeTypeWithUnit(returnType)
		}
		return types.NewFuncType(rt, params...)
	case *ast.TupleType:
		elems := stlslices.Map(t.Elems, func(_ int, p ast.Type) types.Type {
			return a.analyzeType(p)
		})
		return types.NewTupleType(elems...)
	case *ast.ArrayType:
		v, _ := strconv.ParseInt(t.Size.OriginText, 10, 64)
		size := big.NewInt(v)
		elem := a.analyzeType(t.Elem)
		return types.NewArrayType(size, elem)
	case *ast.UnionType:
		elems := stlslices.Map(t.Elems, func(_ int, e ast.Type) types.Type {
			return a.analyzeType(e)
		})
		return types.NewUnionType(elems...)
	case *ast.RefType:
		elem := a.analyzeType(t.Elem)
		return types.NewRefType(t.Mut, elem)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeBuildInIdentType(t *ast.IdentType) types.Type {
	switch t.Name.OriginText {
	case "unit":
		a.reporter.Fatalf(
			t.Name.Position,
			report.Errors.InvalidType,
		)
		return nil
	case "i8":
		return types.I8
	case "i16":
		return types.I16
	case "i32":
		return types.I32
	case "i64":
		return types.I64
	case "u8":
		return types.U8
	case "u16":
		return types.U16
	case "u32":
		return types.U32
	case "u64":
		return types.U64
	case "f32":
		return types.F32
	case "f64":
		return types.F64
	case "bool":
		return types.Bool
	default:
		a.reporter.Fatalf(
			t.Name.Position,
			report.Errors.UnknownIdentifier,
			t.Name.OriginText,
		)
		return nil
	}
}

func (a *Analyzer) analyzeTypeWithUnit(t ast.Type) types.Type {
	if it, ok := t.(*ast.IdentType); ok && it.Name.OriginText == "unit" {
		return types.Unit
	}
	return a.analyzeType(t)
}

func (a *Analyzer) analyzeFuncDecl(f *ast.Func) types.FuncType {
	var returnType types.Type = types.Unit
	if rtAst, ok := f.ReturnType.Value(); ok {
		returnType = a.analyzeTypeWithUnit(rtAst)
	}
	params := stlslices.Map(f.Params, func(_ int, p *ast.ParamDecl) types.Type {
		return a.analyzeType(p.Type)
	})
	return types.NewFuncType(returnType, params...)
}
