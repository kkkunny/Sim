package codegen

import (
	"fmt"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildExpr(expr hir.Expr) cir.Expr {
	switch expr := expr.(type) {
	case hir.Ident:
		return c.buildIdent(expr)
	case *hir.Integer:
		return &cir.IntegerExpr{Value: expr.Value}
	case *hir.Float:
		return &cir.FloatExpr{Value: expr.Value}
	case *hir.Unary:
		return c.buildUnary(expr)
	case *hir.Binary:
		return c.buildBinary(expr)
	case *hir.Func:
		return c.buildFunc(expr)
	case *hir.Call:
		return c.buildCall(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildIdent(expr hir.Ident) cir.Expr {
	name := c.idents[expr].GetName()
	if let, ok := expr.(*hir.Let); ok {
		if stlval.Is[*hir.Func](let.Value) {
			return cir.NewMacroExpr("FUNCEXPR_F", &cir.IdentExpr{Name: name})
		}
	}
	return &cir.IdentExpr{Name: name}
}

func (c *CodeGenerator) buildUnary(expr *hir.Unary) *cir.UnaryExpr {
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

func (c *CodeGenerator) buildBinary(expr *hir.Binary) *cir.BinaryExpr {
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

func (c *CodeGenerator) buildNativeFunc(expr *hir.Func) *cir.FuncExpr {
	params := make([]*cir.Param, len(expr.Params))
	for i, p := range expr.Params {
		pn := fmt.Sprintf("_p%d", i)
		pt := c.buildType(p.Type)
		params[i] = cir.NewParam(pn, pt)
		c.idents[p] = params[i]
	}

	returnType := c.buildType(expr.ReturnType)

	var body optional.Optional[*cir.Block]
	if b, ok := expr.Body.Value(); ok {
		prevBlock, _ := c.builder.CurrentAt()
		c.builder.MoveTo(nil)
		body = optional.Some(c.buildBlock(b))
		c.builder.MoveTo(prevBlock)
	}

	decl := c.builder.BuildFuncDecl("", returnType, params)
	decl.Body = body
	return &cir.FuncExpr{Decl: decl}
}

func (c *CodeGenerator) buildFunc(expr *hir.Func) *cir.MacroExpr {
	f := c.buildNativeFunc(expr)
	return cir.NewMacroExpr("FUNCEXPR_F", &cir.IdentExpr{Name: f.Decl.Name})
}

func (c *CodeGenerator) buildCall(expr *hir.Call) cir.Expr {
	f := c.buildExpr(expr.Func)
	args := stlslices.Map(expr.Args, func(_ int, argExpr hir.Expr) cir.Expr {
		return c.buildExpr(argExpr)
	})
	return cir.NewMacroExpr("FUNCCALL", append([]cir.Expr{f}, args...)...)
}
