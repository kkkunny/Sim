package analyze

import (
	"math/big"
	"strconv"

	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeType(t ast.Type) hir.Type {
	switch t := t.(type) {
	case *ast.IdentType:
		return a.analyzeIdentType(t)
	case *ast.FuncType:
		params := stlslices.Map(t.Params, func(_ int, p ast.Type) hir.Type {
			return a.analyzeType(p)
		})
		var rt hir.Type = types.Unit
		if returnType, ok := t.ReturnType.Value(); ok {
			rt = a.analyzeTypeWithUnit(returnType)
		}
		return types.NewFuncType(rt, params...)
	case *ast.TupleType:
		elems := stlslices.Map(t.Elems, func(_ int, p ast.Type) hir.Type {
			return a.analyzeType(p)
		})
		return types.NewTupleType(elems...)
	case *ast.ArrayType:
		v, _ := strconv.ParseInt(t.Size.OriginText, 10, 64)
		size := big.NewInt(v)
		elem := a.analyzeType(t.Elem)
		return types.NewArrayType(size, elem)
	case *ast.UnionType:
		elems := stlslices.Map(t.Elems, func(_ int, e ast.Type) hir.Type {
			return a.analyzeType(e)
		})
		return types.NewUnionType(elems...)
	case *ast.RefType:
		elem := a.analyzeType(t.Elem)
		return types.NewRefType(t.Mut, elem)
	case *ast.StructType:
		names := set.StdHashSetWithCap[string](uint(len(t.Fields)))
		fields := make([]*types.StructField, len(t.Fields))
		for i, f := range t.Fields {
			name := f.Name.OriginText
			if !names.Add(name) {
				a.reporter.Fatalf(
					f.Name.Position,
					report.Errors.RepeatedIdentifier,
					name,
				)
			}
			fields[i] = &types.StructField{
				Pub:  f.Pub,
				Mut:  f.Mut,
				Name: name,
				Type: a.analyzeType(f.Type),
			}
		}
		return types.NewStructType(fields...)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeIdentType(t *ast.IdentType) hir.Type {
	pkg := a.scope.Root()
	if pkgAst, ok := t.Pkg.Value(); ok {
		pkg, ok = pkg.LookupPkg(pkgAst.OriginText)
		if !ok {
			a.reporter.Fatalf(
				pkgAst.Position,
				report.Errors.UnknownIdentifier,
				pkgAst.OriginText,
			)
		}
	}
	samePkg := pkg == a.scope.Root()

	if ct, ok := pkg.LookupType(t.Name.OriginText); ok {
		if !samePkg && !ct.GetDef().Pub {
			a.reporter.Fatalf(
				t.Name.Position,
				report.Errors.UnknownIdentifier,
				t.Name.OriginText,
			)
		}
		return ct
	} else if !samePkg {
		a.reporter.Fatalf(
			t.Name.Position,
			report.Errors.UnknownIdentifier,
			t.Name.OriginText,
		)
	}
	return a.analyzeBuildInIdentType(t)
}

func (a *Analyzer) analyzeBuildInIdentType(t *ast.IdentType) hir.Type {
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
	case "str":
		return types.Str
	default:
		a.reporter.Fatalf(
			t.Name.Position,
			report.Errors.UnknownIdentifier,
			t.Name.OriginText,
		)
		return nil
	}
}

func (a *Analyzer) analyzeTypeWithUnit(t ast.Type) hir.Type {
	if it, ok := t.(*ast.IdentType); ok && it.Name.OriginText == "unit" {
		return types.Unit
	}
	return a.analyzeType(t)
}

func (a *Analyzer) analyzeFuncDecl(f *ast.Func) types.FuncType {
	var returnType hir.Type = types.Unit
	if rtAst, ok := f.ReturnType.Value(); ok {
		returnType = a.analyzeTypeWithUnit(rtAst)
	}
	params := stlslices.Map(f.Params, func(_ int, p *ast.ParamDecl) hir.Type {
		return a.analyzeType(p.Type)
	})
	return types.NewFuncType(returnType, params...)
}
