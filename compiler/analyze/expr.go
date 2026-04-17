package analyze

import (
	"math/big"
	"strconv"

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
		v, ok := a.scope.Lookup(expr.Name.OriginText)
		if !ok {
			a.reporter.Fatalf(
				expr.Name.Position,
				report.Errors.UnknownIdentifier,
				expr.Name.OriginText,
			)
			return nil
		}
		return &hir.IdentExpr{
			Name: expr.Name.OriginText,
			Type: v.GetType(),
		}
	case *ast.IntegerExpr:
		it, ok := stlslices.Last(expect).(*hir.IntType)
		if !ok {
			it = hir.I32
		}
		v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
		return &hir.IntegerExpr{
			Type:  it,
			Value: big.NewInt(v),
		}
	case *ast.UnaryExpr:
		subExpr := a.analyzeExpr(expr.Expr)
		var op hir.UnaryOp
		switch expr.Op.OriginText {
		case "!":
			op = hir.UnaryOpEnum.Not
		default:
			panic("unreachable")
		}
		return &hir.UnaryExpr{
			Op:   op,
			Expr: subExpr,
		}
	case *ast.BinaryExpr:
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
		default:
			panic("unreachable")
		}
		return &hir.BinaryExpr{
			Op:    op,
			Left:  left,
			Right: right,
		}
	case *ast.FuncExpr:
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

		a.scope = stlval.IgnoreWith(a.scope.Parent())

		return &hir.FuncExpr{
			Params:     params,
			ReturnType: returnType,
			Body:       body,
		}
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
			report.Errors.UnexpectedType,
			expect, vt,
		)
	}
	return v
}
