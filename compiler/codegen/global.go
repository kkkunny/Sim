package codegen

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

func (c *CodeGenerator) genGlobalDecl(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.Let:
		c.genGlobalLetDecl(global)
	}
}

func (c *CodeGenerator) genGlobalLetDecl(l *stmts.Let) {
	if expr, ok := l.Value.(*stmts.Func); ok {
		decl := c.genNativeFuncDecl(expr)
		c.idents[l] = decl
		if l.Name == "main" {
			decl.Name = "sim_main"
		}
		return
	}

	t := c.genType(l.GetType())
	decl := cir.BuildStmt(c.builder, cir.NewVarDecl(t, ""))
	c.idents[l] = decl
}

func (c *CodeGenerator) genGlobalDef(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.Let:
		c.genGlobalLetDef(global)
	case *stmts.TypeDef:
		c.genTypeDef(global)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genGlobalLetDef(l *stmts.Let) {
	decl := c.idents[l]

	if expr, ok := l.Value.(*stmts.Func); ok {
		funcDecl := decl.(*cir.FuncDecl)
		if b, ok := expr.Body.Value(); ok {
			prevFunc := c.currentFunc
			c.currentFunc = expr
			funcDecl.Body = optional.Some(c.genFuncBlock(b, nil))
			c.currentFunc = prevFunc
		}
		return
	}

	varDecl := decl.(*cir.VarDecl)
	varDecl.Value = optional.Some(c.genExpr(l.Value))
}

func (c *CodeGenerator) genTypeDef(global *stmts.TypeDef) {
	underlying := c.genType(global.Type.GetUnderlying())
	def := cir.BuildStmt(c.builder, cir.NewTypedef(underlying, ""))
	c.typeCache[global.Type.GetName()] = cir.NewAliasType(def)
}
