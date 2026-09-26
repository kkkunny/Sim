package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genExpr(expr locals.Expr) llvm.AnyValue {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		return c.genIdentExpr(expr)
	case *locals.Integer:
		return c.genInteger(expr)
	case *locals.Float:
		return c.genFloat(expr)
	case *locals.Boolean:
		return c.ctx.LLVM().ConstBool(expr.Value)
	case *locals.String:
		return c.genString(expr)
	case *locals.Call:
		return c.genCall(expr)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的表达式 %s（%T）", expr, expr))
	}
}

func (c *CodeGenerator) genIdentExpr(expr *locals.IdentExpr) llvm.AnyValue {
	ident, ok := c.ctx.idents[expr.Define]
	if !ok {
		panic(fmt.Errorf("llgen: 未找到符号 %s", expr.Define.GetName()))
	}
	if !ident.Local.IsNil() {
		return c.builder.Load[llvm.DynT](ident.Local, c.genType(expr.GetType()).DynType(), "")
	}
	panic(fmt.Errorf("llgen: 暂不支持的非局部标识符 %s（C4/F3）", ident.Name))
}

func (c *CodeGenerator) genInteger(expr *locals.Integer) llvm.AnyValue {
	it, ok := c.genType(expr.GetType()).(llvm.IntType)
	if !ok {
		panic(fmt.Errorf("llgen: 整数字面量的类型 %s 不是整数类型", expr.GetType()))
	}
	return c.ctx.LLVM().ConstIntOfString(it, expr.Value.String(), 10)
}

func (c *CodeGenerator) genFloat(expr *locals.Float) llvm.AnyValue {
	ft, ok := c.genType(expr.GetType()).(llvm.FloatType)
	if !ok {
		panic(fmt.Errorf("llgen: 浮点字面量的类型 %s 不是浮点类型", expr.GetType()))
	}
	v, _ := expr.Value.Float64()
	return c.ctx.LLVM().ConstFloat(ft, v)
}

func (c *CodeGenerator) genString(expr *locals.String) llvm.AnyValue {
	c.strCount++
	data := c.ctx.LLVM().ConstString(expr.Value, true)
	g := c.module.NewGlobalConst(fmt.Sprintf("_str.%d", c.strCount), data)
	ptr := c.ctx.LLVM().ConstGEP(
		c.ctx.LLVM().Int(8), g, true,
		c.ctx.LLVM().Int(32).Const(0), c.ctx.LLVM().Int(32).Const(0),
	)
	st, ok := c.genType(expr.GetType()).(llvm.StructType)
	if !ok {
		panic(fmt.Errorf("llgen: 字符串字面量的类型 %s 不是结构体", expr.GetType()))
	}
	return c.ctx.LLVM().ConstNamedStruct(st, ptr, c.ctx.LLVM().Int(64).Const(uint64(len(expr.Value))))
}

func (c *CodeGenerator) genCall(expr *locals.Call) llvm.AnyValue {
	// 直接调用：被调方是函数符号
	if identExpr, ok := expr.Func.(*locals.IdentExpr); ok {
		if ident, ok := c.ctx.idents[identExpr.Define]; ok && ident.Local.IsNil() {
			ft, ok := expr.Func.GetType().(types.FuncType)
			if !ok {
				panic(fmt.Errorf("llgen: 被调用符号 %s 不是函数类型", ident.Name))
			}
			fn := c.getFunction(ident.Name, c.genNativeFuncType(ft))
			args := stlslices.Map(expr.Args, func(_ int, e locals.Expr) llvm.AnyValue {
				return c.genExpr(e)
			})
			return c.builder.Call[llvm.DynT](fn, args, "")
		}
	}
	panic(fmt.Errorf("llgen: 暂不支持的调用形式（%T，D13/F4）", expr.Func))
}

func (c *CodeGenerator) getFunction(name string, sig llvm.FnType) ir.Function {
	if fn, ok := c.module.GetFunction(name); ok {
		return fn
	}
	return c.module.NewFunction(name, sig)
}
