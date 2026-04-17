package codegen

import (
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
		return &cir.UnaryExpr{
			Op:   c.buildUnaryOp(expr.Op),
			Expr: c.buildExpr(expr.Expr),
		}
	case *hir.BinaryExpr:
		return &cir.BinaryExpr{
			Op:    c.buildBinaryOp(expr.Op),
			Left:  c.buildExpr(expr.Left),
			Right: c.buildExpr(expr.Right),
		}
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildUnaryOp(op hir.UnaryOp) cir.UnaryOp {
	switch op {
	case hir.UnaryOpEnum.Not:
		return cir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildBinaryOp(op hir.BinaryOp) cir.BinaryOp {
	switch op {
	case hir.BinaryOpEnum.Add:
		return cir.BinaryOpEnum.Add
	case hir.BinaryOpEnum.Sub:
		return cir.BinaryOpEnum.Sub
	case hir.BinaryOpEnum.Mul:
		return cir.BinaryOpEnum.Mul
	case hir.BinaryOpEnum.Quo:
		return cir.BinaryOpEnum.Quo
	case hir.BinaryOpEnum.Rem:
		return cir.BinaryOpEnum.Rem
	case hir.BinaryOpEnum.And:
		return cir.BinaryOpEnum.And
	case hir.BinaryOpEnum.Or:
		return cir.BinaryOpEnum.Or
	case hir.BinaryOpEnum.Xor:
		return cir.BinaryOpEnum.Xor
	default:
		panic("unreachable")
	}
}
