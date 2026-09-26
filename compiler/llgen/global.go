package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genTypeDecl(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		c.genCustomTypeDecl(global)
	}
}

func (c *CodeGenerator) genCustomTypeDecl(global *globals.TypeDef) {
	switch global.Underlying.(type) {
	case types.TupleType, types.ArrayType, types.UnionType, types.FuncType, types.RefType, types.StructType:
	default:
		return
	}
	// 预声明 opaque named struct，支持递归/互递归类型（body 由 B 系列任务填充）
	name := stableName(c.pkg, global.Name)
	c.ctx.typeCache[global] = c.ctx.NamedStruct(name)
}

func (c *CodeGenerator) genTypeDef(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		c.genCustomTypeDef(global)
	}
}

func (c *CodeGenerator) genCustomTypeDef(global *globals.TypeDef) llvm.AnyType {
	if t, ok := c.ctx.typeCache[global]; ok {
		// 预声明过的聚合类型：body 由 B 系列任务填充
		// TODO(B6/B7): t.(llvm.StructType).SetBody(...)
		return t
	}
	// 标量别名：直接透传底层类型
	t := c.genType(global.Underlying)
	c.ctx.typeCache[global] = t
	return t
}

func (c *CodeGenerator) genGlobalValue(global globals.Global) {
	switch global := global.(type) {
	case *locals.Let:
		c.genGlobalLet(global)
	}
}

// genFuncDecls 函数符号声明子遍（D13）：在生成任何函数体之前登记全部函数符号，
// 使同包前向引用（helper 定义在调用点之后）与互递归可用。非函数全局跳过。
func (c *CodeGenerator) genFuncDecls() {
	for _, g := range c.pkg.Globals {
		l, ok := g.(*locals.Let)
		if !ok || l.Mut || !stlval.Is[types.FuncType](l.GetType()) {
			continue
		}
		if v, ok := l.Value.Value(); ok {
			if stlval.Is[*locals.Func](v) {
				c.genGlobalFuncDecl(l)
			}
		} else {
			c.genExternalFuncDecl(l)
		}
	}
}

func (c *CodeGenerator) genGlobalLet(l *locals.Let) {
	if !l.Mut && stlval.Is[types.FuncType](l.GetType()) {
		if l.Value.IsSome() && stlval.Is[*locals.Func](l.Value.MustValue()) {
			c.genGlobalFunc(l, l.Value.MustValue().(*locals.Func))
			return
		} else if l.Value.IsNone() {
			c.genExternalFunc(l)
			return
		}
	}
	panic(fmt.Errorf("llgen: 暂不支持全局变量 %s（C4）", l.Name))
}

func (c *CodeGenerator) genGlobalFunc(l *locals.Let, expr *locals.Func) {
	c.genGlobalFuncDecl(l)
	body, ok := expr.Body.Value()
	if !ok {
		return
	}
	ident := c.ctx.idents[l]
	decl, ok := c.module.GetFunction(ident.Name)
	if !ok {
		panic(fmt.Errorf("llgen: 函数 %s 的符号未在模块中登记（%s）", l.Name, ident.Name))
	}
	c.genFuncBody(decl, expr, body, nil)
}

// genGlobalFuncDecl 建函数符号并登记（不生成函数体）；已登记则直接复用
func (c *CodeGenerator) genGlobalFuncDecl(l *locals.Let) {
	if _, ok := c.ctx.idents[l]; ok {
		return
	}
	name := stableName(c.pkg, l.Name)
	internal := !l.Pub
	switch {
	case l.Name == "main":
		name = "sim_main"
		internal = true
	case l.ExternalName.IsSome():
		name = l.ExternalName.MustValue()
		internal = false
	}
	decl := c.module.NewFunction(name, c.genNativeFuncType(l.GetType().(types.FuncType)))
	if internal {
		decl.SetLinkage(llvm.LinkageInternal)
	}
	c.ctx.idents[l] = &Ident{Name: name}
}

func (c *CodeGenerator) genExternalFunc(l *locals.Let) {
	c.genExternalFuncDecl(l)
}

// genExternalFuncDecl 建外部函数声明并登记；已登记则直接复用
func (c *CodeGenerator) genExternalFuncDecl(l *locals.Let) {
	if _, ok := c.ctx.idents[l]; ok {
		return
	}
	decl := c.module.NewFunction(l.ExternalName.MustValue(), c.genNativeFuncType(l.GetType().(types.FuncType)))
	c.ctx.idents[l] = &Ident{Name: decl.Name(), ExternalFunc: true}
}
