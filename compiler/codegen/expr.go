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

func (c *CodeGenerator) buildBinaryExpr(expr *ast.BinaryExpr) *cir.BinaryExpr {
	return &cir.BinaryExpr{
		Op:    c.buildBinaryOp(expr.Op),
		Left:  c.buildExpr(expr.Left),
		Right: c.buildExpr(expr.Right),
	}
}

func (c *CodeGenerator) buildBinaryOp(op token.Token) cir.BinaryOp {
	switch op.Kind {
	case token.KindEnum.Add:
		return cir.BinaryOpEnum.Add
	case token.KindEnum.Sub:
		return cir.BinaryOpEnum.Sub
	case token.KindEnum.Mul:
		return cir.BinaryOpEnum.Mul
	case token.KindEnum.Quo:
		return cir.BinaryOpEnum.Quo
	case token.KindEnum.Rem:
		return cir.BinaryOpEnum.Rem
	default:
		panic("unreachable")
	}
}
