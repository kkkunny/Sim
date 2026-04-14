package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
)

func (c *CodeGenerator) genExpr(expr ast.Expr) {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		c.genIdentExpr(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genIdentExpr(expr *ast.IdentExpr) {
	c.buf.WriteString(expr.Name.OriginText)
}
