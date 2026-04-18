package codegen

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildExpr(expr hir.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *hir.IdentExpr:
		return &cir.IdentExpr{Name: expr.Name}
	case *hir.IntegerExpr:
		return &cir.IntegerExpr{Value: expr.Value}
	case *hir.FloatExpr:
		return &cir.FloatExpr{Value: expr.Value}
	case *hir.UnaryExpr:
		return c.buildUnaryOp(expr)
	case *hir.BinaryExpr:
		return c.buildBinaryExpr(expr)
	case *hir.FuncExpr:
		return c.buildFuncExpr(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildUnaryOp(expr *hir.UnaryExpr) *cir.UnaryExpr {
	var op cir.UnaryOp
	switch expr.Op {
	case hir.UnaryOpEnum.Not:
		op = cir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}
	return &cir.UnaryExpr{
		Op:   op,
		Expr: c.buildExpr(expr.Expr),
	}
}

func (c *CodeGenerator) buildBinaryExpr(expr *hir.BinaryExpr) *cir.BinaryExpr {
	var op cir.BinaryOp
	switch expr.Op {
	case hir.BinaryOpEnum.Add:
		op = cir.BinaryOpEnum.Add
	case hir.BinaryOpEnum.Sub:
		op = cir.BinaryOpEnum.Sub
	case hir.BinaryOpEnum.Mul:
		op = cir.BinaryOpEnum.Mul
	case hir.BinaryOpEnum.Quo:
		op = cir.BinaryOpEnum.Quo
	case hir.BinaryOpEnum.Rem:
		op = cir.BinaryOpEnum.Rem
	case hir.BinaryOpEnum.And:
		op = cir.BinaryOpEnum.And
	case hir.BinaryOpEnum.Or:
		op = cir.BinaryOpEnum.Or
	case hir.BinaryOpEnum.Xor:
		op = cir.BinaryOpEnum.Xor
	default:
		panic("unreachable")
	}
	return &cir.BinaryExpr{
		Op:    op,
		Left:  c.buildExpr(expr.Left),
		Right: c.buildExpr(expr.Right),
	}
}

func (c *CodeGenerator) buildFuncExpr(expr *hir.FuncExpr) *cir.FuncExpr {
	var body optional.Optional[*cir.Block]
	if b, ok := expr.Body.Value(); ok {
		body = optional.Some(c.buildBlock(b))
	}
	returnType := c.buildType(expr.ReturnType)
	params := make([]*cir.ParamDecl, len(expr.Params))
	for i, p := range expr.Params {
		params[i] = &cir.ParamDecl{
			Name: p.Name,
			Type: c.buildType(p.Type),
		}
	}
	decl := c.builder.BuildFuncDecl("", returnType, params)
	decl.Body = body
	return &cir.FuncExpr{Decl: decl}
}
