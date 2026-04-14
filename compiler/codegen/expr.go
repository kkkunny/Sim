package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
)

func (c *CodeGenerator) buildExpr(expr ast.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return c.buildIdentExpr(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildIdentExpr(expr *ast.IdentExpr) *cir.IdentExpr {
	return &cir.IdentExpr{Name: expr.Name.OriginText}
}
