package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
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
	case types.StructType, types.TupleType, types.UnionType:
	default:
		// 数组/引用/函数别名无需预声明 named struct（B7）：
		// 引用经 ptr 天然断开递归；数组映射为 [N x T]；函数值映射为胖结构 {ptr, ptr}
		return
	}
	// 预声明 opaque named struct，支持递归/互递归聚合（body 由 genCustomTypeDef 填充）
	name := stableName(c.pkg, global.Name)
	c.ctx.typeCache[global] = c.ctx.NamedStruct(name)
}

func (c *CodeGenerator) genTypeDef(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		c.genCustomTypeDef(global)
	}
}

// genCustomTypeDef 填充/映射自定义类型（B7）：预声明过的聚合填充 body；
// 标量/数组/引用/函数别名直接映射底层 LLVM 类型并登记（transparent）。
func (c *CodeGenerator) genCustomTypeDef(global *globals.TypeDef) llvm.AnyType {
	if t, ok := c.ctx.typeCache[global]; ok {
		// 预声明的聚合：填充 body（已填充则跳过）；透明别名：直接返回
		if st, ok := t.(llvm.StructType); ok && st.IsOpaque() {
			st.SetBody(c.customAggregateElems(global.Underlying), false)
		}
		return t
	}
	t := c.genType(global.Underlying)
	c.ctx.typeCache[global] = t
	return t
}

// customAggregateElems 预声明的自定义聚合 underlying → LLVM 成员类型
func (c *CodeGenerator) customAggregateElems(underlying hir.Type) []llvm.AnyType {
	switch underlying := underlying.(type) {
	case types.StructType:
		return c.genTypes(stlslices.Map(underlying.GetFields(), func(_ int, f *types.StructField) hir.Type {
			return f.Type
		}))
	case types.TupleType:
		return c.genTypes(underlying.GetElems())
	case types.UnionType:
		elems := c.unionMemberTypes(underlying)
		if !unionHasPayload(elems) {
			return nil // 全零尺寸 union → {}
		}
		payload, _ := c.genUnionPayload(elems)
		return []llvm.AnyType{c.ctx.LLVM().Int(8), payload}
	default:
		panic(fmt.Errorf("llgen: 预声明的自定义类型 %T 不是聚合类型", underlying))
	}
}

// genGlobalValue 全局值生成（变量/函数定义）
func (c *CodeGenerator) genGlobalValue(global globals.Global) {
	switch global := global.(type) {
	case *locals.Let:
		c.genGlobalLet(global)
	}
}

// genGlobalVarDecls 全局变量符号声明子遍（C4）：在生成任何函数体之前登记全部全局变量的
// 符号与名称，使函数体可以前向引用同包后置声明的全局（函数符号见 genFuncDecls）。
// 链接性与初始化器在 genGlobalVar 中补齐。
func (c *CodeGenerator) genGlobalVarDecls() {
	for _, g := range c.pkg.Globals {
		l, ok := g.(*locals.Let)
		if !ok || (!l.Mut && stlval.Is[types.FuncType](l.GetType())) {
			continue
		}
		if _, ok := c.ctx.idents[l]; ok {
			continue
		}
		name, internal := c.globalVarName(l)
		decl := c.getGlobalVar(name, c.genType(l.GetType()))
		if internal {
			decl.SetLinkage(llvm.LinkageInternal)
		}
		c.ctx.idents[l] = &Ident{Name: name}
	}
}

// globalVarName 全局变量符号名与是否 internal 链接（ExternalName → 原名 external；
// Pub → stableName external；否则 stableName + internal）
func (c *CodeGenerator) globalVarName(l *locals.Let) (string, bool) {
	if ext, ok := l.ExternalName.Value(); ok {
		return ext, false
	}
	return stableName(c.pkg, l.Name), !l.Pub
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
	c.genGlobalVar(l)
}

// genGlobalVar 全局变量定义（C4）：ExternalName → 原名 external；Pub → stableName external；
// 否则 stableName + internal。有值 → 常量初始化；包内无值 → 零初始化定义（C tentative definition）；
// @extern 无值 → 纯声明。
func (c *CodeGenerator) genGlobalVar(l *locals.Let) {
	name, internal := c.globalVarName(l)
	t := c.genType(l.GetType())
	g := c.getGlobalVar(name, t)
	if internal {
		g.SetLinkage(llvm.LinkageInternal)
	}
	c.ctx.idents[l] = &Ident{Name: name}

	if v, ok := l.Value.Value(); ok {
		// LLVM 全局初始化器必须是常量；C 后端同样只能处理常量表达式
		cv, ok := c.genConstExpr(v)
		if !ok {
			panic(fmt.Errorf("llgen: 暂不支持非常量全局初始化 %s = %s（%T）", l.Name, v, v))
		}
		g.SetInitializer(cv)
	} else if l.ExternalName.IsNone() {
		g.SetInitializer(c.genZeroValue(t))
	}
}

// getGlobalVar 获取全局变量句柄；本模块尚无该符号时按类型创建外部声明
// （跨包引用与同包后向引用都走这里，符号名与类型由调用方保证一致）
func (c *CodeGenerator) getGlobalVar(name string, t llvm.AnyType) ir.Global {
	if g, ok := c.module.GetGlobal(name); ok {
		return g
	}
	return c.module.NewGlobal(name, t)
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
	c.genFuncBody(decl, expr, body, 0, nil)
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
	c.ctx.idents[l] = &Ident{Name: name, FuncSymbol: true}
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
	c.ctx.idents[l] = &Ident{Name: decl.Name(), ExternalFunc: true, FuncSymbol: true}
}
