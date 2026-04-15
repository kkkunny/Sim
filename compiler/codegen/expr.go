package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/token"
)

func (c *CodeGenerator) buildExpr(expr ast.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return c.buildIdentExpr(expr)
	case *ast.IntegerExpr:
		return c.buildIntegerExpr(expr)
	case *ast.UnaryExpr:
		return c.buildUnaryExpr(expr)
	case *ast.BinaryExpr:
		return c.buildBinaryExpr(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildIdentExpr(expr *ast.IdentExpr) *cir.IdentExpr {
	return &cir.IdentExpr{Name: expr.Name.OriginText}
}

func (c *CodeGenerator) buildIntegerExpr(expr *ast.IntegerExpr) *cir.IntegerExpr {
	return &cir.IntegerExpr{Value: expr.Value.OriginText}
}

func (c *CodeGenerator) buildUnaryExpr(expr *ast.UnaryExpr) *cir.UnaryExpr {
	var op cir.UnaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Not:
		op = cir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}

	return &cir.UnaryExpr{
		Op:   op,
		Expr: c.buildExpr(expr.Expr),
	}
}

func (c *CodeGenerator) buildBinaryExpr(expr *ast.BinaryExpr) *cir.BinaryExpr {
	var op cir.BinaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Add:
		op = cir.BinaryOpEnum.Add
	case token.KindEnum.Sub:
		op = cir.BinaryOpEnum.Sub
	case token.KindEnum.Mul:
		op = cir.BinaryOpEnum.Mul
	case token.KindEnum.Quo:
		op = cir.BinaryOpEnum.Quo
	case token.KindEnum.Rem:
		op = cir.BinaryOpEnum.Rem
	case token.KindEnum.And:
		op = cir.BinaryOpEnum.And
	case token.KindEnum.Or:
		op = cir.BinaryOpEnum.Or
	case token.KindEnum.Xor:
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
