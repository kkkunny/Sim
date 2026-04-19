package analyze

import (
	"fmt"
	"math/big"
	"strconv"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeExpr(expr ast.Expr, expect ...hir.Type) hir.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return a.analyseIdentExpr(expr)
	case *ast.Integer:
		return a.analyseInteger(expr, expect...)
	case *ast.Unary:
		return a.analyseUnary(expr)
	case *ast.Binary:
		return a.analyseBinary(expr, expect...)
	case *ast.Func:
		return a.analyseFunc(expr)
	case *ast.Call:
		return a.analyseCall(expr)
	case *ast.Tuple:
		return a.analyzeTuple(expr, expect...)
	case *ast.Index:
		return a.analyzeIndex(expr)
	case *ast.Array:
		return a.analyzeArray(expr, expect...)
	default:
		panic("unreachable")
	}
}

// 期待类型，两个类型必须完全相同
func (a *Analyzer) expectTypeExpr(expr ast.Expr, expect hir.Type) hir.Expr {
	v := a.analyzeExpr(expr, expect)
	if vt := v.GetType(); !vt.Equal(expect) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpression,
			expect, vt,
		)
	}
	return v
}

// 期待类型，属于制定的类型类
func expectTypeExpr[TYPE hir.Type](a *Analyzer, expr ast.Expr) hir.Expr {
	v := a.analyzeExpr(expr)
	if vt := v.GetType(); !stlval.Is[TYPE](vt) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpressionType,
			fmt.Sprintf("%T", stlval.Default[TYPE]()), vt,
		)
	}
	return v
}

func (a *Analyzer) analyseIdentExpr(expr *ast.IdentExpr) *hir.IdentExpr {
	v, ok := a.scope.Lookup(expr.Name.OriginText)
	if !ok {
		a.reporter.Fatalf(
			expr.Name.Position,
			report.Errors.UnknownIdentifier,
			expr.Name.OriginText,
		)
		return nil
	}
	return hir.NewIdentExpr(v)
}

func (a *Analyzer) analyseInteger(expr *ast.Integer, expect ...hir.Type) hir.Expr {
	var t hir.Type
	it, ok := stlslices.Last(expect).(hir.IntegerType)
	if ok {
		t = it
	} else {
		ft, ok := stlslices.Last(expect).(*hir.FloatType)
		if ok {
			t = ft
		} else {
			t = hir.I32
		}
	}
	v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
	if stlval.Is[hir.IntegerType](t) {
		return &hir.Integer{
			Type:  t,
			Value: big.NewInt(v),
		}
	} else {
		return &hir.Float{
			Type:  t,
			Value: big.NewFloat(float64(v)),
		}
	}
}

func (a *Analyzer) analyseUnary(expr *ast.Unary) *hir.Unary {
	subExpr := a.analyzeExpr(expr.Expr)
	var op hir.UnaryOp
	switch expr.Op.OriginText {
	case "!":
		op = hir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}
	return &hir.Unary{
		Op:   op,
		Expr: subExpr,
	}
}

func (a *Analyzer) analyseBinary(expr *ast.Binary, expect ...hir.Type) *hir.Binary {
	left := a.analyzeExpr(expr.Left, expect...)
	right := a.expectTypeExpr(expr.Right, left.GetType())
	var op hir.BinaryOp
	switch expr.Op.OriginText {
	case "+":
		op = hir.BinaryOpEnum.Add
	case "-":
		op = hir.BinaryOpEnum.Sub
	case "*":
		op = hir.BinaryOpEnum.Mul
	case "/":
		op = hir.BinaryOpEnum.Quo
	case "%":
		op = hir.BinaryOpEnum.Rem
	case "&":
		op = hir.BinaryOpEnum.And
	case "|":
		op = hir.BinaryOpEnum.Or
	case "^":
		op = hir.BinaryOpEnum.Xor
	case "=":
		op = hir.BinaryOpEnum.Assign
	default:
		panic("unreachable")
	}
	return &hir.Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (a *Analyzer) analyseFunc(expr *ast.Func) *hir.Func {
	scope := hir.NewBlockScope(a.scope)
	a.scope = scope

	var params []*hir.Param
	for _, p := range expr.Params {
		paramType := a.analyzeType(p.Type)
		params = append(params, &hir.Param{
			Name: p.Name.OriginText,
			Type: paramType,
		})
	}

	var returnType hir.Type = hir.Unit
	if rtAst, ok := expr.ReturnType.Value(); ok {
		returnType = a.analyzeType(rtAst)
	}

	ft := hir.NewFuncType(returnType, stlslices.Map(params, func(i int, p *hir.Param) hir.Type {
		return p.Type
	})...)
	scope.SetFuncType(ft)

	for _, p := range params {
		a.scope.AddValue(p)
	}

	var body optional.Optional[*hir.Block]
	if b, ok := expr.Body.Value(); ok {
		body = optional.Some(a.analyzeBlock(b))
	}

	externalVars := stlslices.DiffTo(a.scope.UsedValues(), stlmaps.Values(a.scope.Values()))

	a.scope = stlval.IgnoreWith(a.scope.Parent())

	return &hir.Func{
		Params:                params,
		ReturnType:            returnType,
		Body:                  body,
		UsedExternalVariables: externalVars,
	}
}

func (a *Analyzer) analyseCall(expr *ast.Call) *hir.Call {
	f := expectTypeExpr[*hir.FuncType](a, expr.Func)
	ft := f.GetType().(*hir.FuncType)
	if len(ft.Params) != len(expr.Args) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.InsufficientArguments,
			len(ft.Params), len(expr.Args),
		)
	}
	args := stlslices.Map(expr.Args, func(i int, expr ast.Expr) hir.Expr {
		return a.analyzeExpr(expr, ft.Params[i])
	})
	return &hir.Call{
		Func: f,
		Args: args,
	}
}

func (a *Analyzer) analyzeTuple(expr *ast.Tuple, expect ...hir.Type) hir.Expr {
	var expectTuple bool
	expectElemTypes := make([]hir.Type, len(expr.Elems))
	if len(expect) > 0 {
		if tt, ok := stlslices.Last(expect).(*hir.TupleType); ok && len(expr.Elems) == len(tt.Elems) {
			expectElemTypes = tt.Elems
			expectTuple = true
		} else if len(expr.Elems) == 1 {
			expectElemTypes[0] = stlslices.Last(expect)
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) hir.Expr {
		var elemExpect []hir.Type
		if et := expectElemTypes[i]; et != nil {
			elemExpect = []hir.Type{et}
		}
		return a.analyzeExpr(e, elemExpect...)
	})

	if len(elems) == 1 && !expectTuple {
		return elems[0]
	}

	return hir.NewTuple(elems...)
}

func (a *Analyzer) analyzeIndex(expr *ast.Index) hir.Expr {
	from := a.analyzeExpr(expr.From)
	ft := from.GetType()

	if stlval.Is[*hir.TupleType](ft) {
		index := expectTypeExpr[hir.IntegerType](a, expr.Index)
		indexValue, ok := index.(*hir.Integer)
		if !ok {
			a.reporter.Fatalf(
				expr.Index.Position(),
				report.Errors.ExpectedIntegerConstant,
			)
		}
		return hir.NewTupleIndex(from, indexValue.Value)
	}

	at, ok := ft.(*hir.ArrayType)
	if !ok {
		a.reporter.Fatalf(
			expr.From.Position(),
			report.Errors.UnexpectedExpressionType,
			fmt.Sprintf("%T", hir.ArrayType{}), at,
		)
	}
	index := expectTypeExpr[hir.IntegerType](a, expr.Index)
	return hir.NewArrayIndex(from, index)
}

func (a *Analyzer) analyzeArray(expr *ast.Array, expect ...hir.Type) *hir.Array {
	var expectElemType hir.Type
	if len(expect) > 0 {
		if at, ok := stlslices.Last(expect).(*hir.ArrayType); ok && at.Size.Int64() == int64(len(expr.Elems)) {
			expectElemType = at.Elem
		}
	}

	if len(expr.Elems) == 0 {
		if expectElemType == nil {
			a.reporter.Fatalf(
				expr.Position(),
				report.Errors.TypeLoss,
			)
		}
		return hir.NewArray(expectElemType)
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) hir.Expr {
		if i == 0 {
			var elemExpect []hir.Type
			if expectElemType != nil {
				elemExpect = []hir.Type{expectElemType}
			}
			v := a.analyzeExpr(e, elemExpect...)
			expectElemType = v.GetType()
			return v
		} else {
			return a.expectTypeExpr(e, expectElemType)
		}
	})

	return hir.NewArray(hir.NewArrayType(big.NewInt(int64(len(elems))), expectElemType), elems...)
}
