package analyze

import (
	"math/big"
	"strconv"

	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeExpr(expr ast.Expr) hir.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		typ, ok := a.scope.Types[expr.Name.OriginText]
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
			Type: typ,
		}
	case *ast.IntegerExpr:
		v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
		return &hir.IntegerExpr{
			Value: big.NewInt(v),
			Type:  hir.I32,
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
		left := a.analyzeExpr(expr.Left)
		right := a.analyzeExpr(expr.Right)
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
		a.scope = &hir.Scope{Parent: a.scope, Types: make(map[string]hir.Type)}

		var params []*hir.ParamDecl
		for _, p := range expr.Params {
			paramType := a.analyzeType(p.Type)
			params = append(params, &hir.ParamDecl{
				Name: p.Name.OriginText,
				Type: paramType,
			})
			a.scope.Types[p.Name.OriginText] = paramType
		}

		var returnType hir.Type = hir.Unit
		if rtAst, ok := expr.ReturnType.Value(); ok {
			returnType = a.analyzeType(rtAst)
		}

		var body optional.Optional[*hir.Block]
		if b, ok := expr.Body.Value(); ok {
			body = optional.Some(a.analyzeBlock(b))
		}

		return &hir.FuncExpr{
			Params:     params,
			ReturnType: returnType,
			Body:       body,
		}
	default:
		panic("unreachable")
	}
}
