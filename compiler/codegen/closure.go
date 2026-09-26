package codegen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// closureInfo 函数字面量对应的内部函数与捕获布局（§4.4）
type closureInfo struct {
	fn       ir.Function // 内部函数；有捕获时为包装函数（首参 ptr ctx）
	captures []hir.Ident // 捕获变量（按 ctx 字段顺序 _f1.._fn）
	ctxT     llvm.StructType
}

// buildClosure 创建（或复用）函数字面量的内部函数并生成函数体：
//   - 无捕获：普通 internal 函数（static），值为 {fnptr, null}；
//   - 有捕获：ctx named struct（字段 _f1.._fn）+ 首参 ctx ptr 的包装函数，
//     函数体内捕获变量解析为 ctx 字段地址（F2/F7）。
//
// 函数体在定义点生成一次（同一字面量只生成一次），生成期间保存并恢复调用点的
// 发射状态（定义点位于外层函数体中间）。
func (c *CodeGenerator) buildClosure(expr *locals.Func) *closureInfo {
	if info, ok := c.closures[expr]; ok {
		return info
	}
	info := &closureInfo{captures: expr.CaptureVariables}
	c.ctx.closureCount++
	name := fmt.Sprintf("_closure.%d", c.ctx.closureCount)
	funcT, ok := asFuncType(expr.GetType())
	if !ok {
		panic(c.ice("type %s of function literal is not a func type", expr.GetType()))
	}
	if len(info.captures) == 0 {
		info.fn = c.module.NewFunction(name, c.genNativeFuncType(funcT))
	} else {
		fields := make([]llvm.AnyType, len(info.captures))
		for i, cv := range info.captures {
			fields[i] = c.genType(cv.GetType())
		}
		info.ctxT = c.ctx.NamedStructWithBody(fmt.Sprintf("_closure_ctx.%d", c.ctx.closureCount), fields...)
		info.fn = c.module.NewFunction(name, c.genCtxFuncType(funcT))
	}
	info.fn.SetLinkage(llvm.LinkageInternal)
	c.closures[expr] = info // 先登记再生成函数体，避免递归生成

	body, ok := expr.Body.Value()
	if !ok {
		return info
	}

	paramOffset := 0
	if len(info.captures) > 0 {
		paramOffset = 1
	}
	prevCapture := c.captureVars
	defer func() {
		c.captureVars = prevCapture
	}()
	c.genFuncBody(info.fn, expr, body, paramOffset, func() {
		if len(info.captures) == 0 {
			return
		}
		// 捕获解析：ctx 字段地址（GEP），优先级高于 c.ctx.idents（F7）
		ctxPtr := info.fn.Param(0).Dyn().MustAs[llvm.PtrT]()
		captureVars := make(map[hir.Ident]llvm.Value[llvm.PtrT], len(info.captures))
		for i, cv := range info.captures {
			captureVars[cv] = c.builder.GEP(info.ctxT, ctxPtr,
				c.gepPath(c.ctx.LLVM().Int(32).Const(uint64(i))), "")
		}
		c.captureVars = captureVars
	})
	return info
}

// genClosureValue 函数字面量作为值（F2/F3）：
// 无捕获 → 常量 {fnptr, null}；有捕获 → 定义点求值捕获值存入 ctx，值为 {fnptr, &ctx}。
func (c *CodeGenerator) genClosureValue(expr *locals.Func) llvm.AnyValue {
	info := c.buildClosure(expr)
	if len(info.captures) == 0 {
		return c.packNullFuncValue(info.fn.AsValue())
	}
	return c.packFuncValue(info.fn.AsValue(), c.buildClosureCtx(info))
}

// buildClosureCtx 定义点构造闭包 ctx（§4.4）：入口块 alloca → 逐字段 store 捕获值。
// 捕获按值语义：每个捕获只求值一次（在外层上下文中求值，可为外层闭包的 ctx 字段），
// 之后闭包的所有调用共享同一 ctx 实例（可变捕获写回对后续调用可见）。
func (c *CodeGenerator) buildClosureCtx(info *closureInfo) llvm.Value[llvm.PtrT] {
	slot := c.allocaEntry(info.ctxT, "")
	i32 := c.ctx.LLVM().Int(32)
	for i, cv := range info.captures {
		field := c.builder.GEP(info.ctxT, slot, c.gepPath(i32.Const(uint64(i))), "")
		c.builder.Store(c.genExpr(locals.NewIdentExpr(cv)), field)
	}
	return slot
}

// genFuncSymbolValue 全局函数/外部函数作为值（F3）：{fnptr, null}
func (c *CodeGenerator) genFuncSymbolValue(ident *Ident, funcT types.FuncType) llvm.AnyValue {
	fn := c.getFunction(ident.Name, c.genNativeFuncType(funcT))
	return c.packNullFuncValue(fn.AsValue())
}

// packNullFuncValue 无捕获函数值：常量 { fnptr, null }
func (c *CodeGenerator) packNullFuncValue(fn llvm.AnyValue) llvm.AnyValue {
	ptrT := c.ctx.LLVM().Ptr(0)
	return c.ctx.LLVM().ConstStruct(false, fn, ptrT.Null())
}

// packFuncValue 构造带 ctx 的胖函数值 { fn, ctx }：以零值为基逐字段 InsertValue。
// ctx 为运行时地址，不能走 ConstStruct（常量构造要求全部元素为常量）。
func (c *CodeGenerator) packFuncValue(fn, ctxPtr llvm.AnyValue) llvm.AnyValue {
	fatT := c.genFatFuncType()
	v1 := c.builder.InsertValue[llvm.StructT](fatT.Zero(), fn, []uint32{0}, "")
	return c.builder.InsertValue[llvm.StructT](v1, ctxPtr, []uint32{1}, "")
}

// genGetBind 方法绑定（F6）：复刻旧 genGetBind 的 HIR 层包装，再走闭包生成。
// 静态绑定（外部/可变/无 self 参数）直接取函数符号值；否则构造
// `() -> ... { return bind(self, params...) }` 的函数字面量，self 为捕获变量。
func (c *CodeGenerator) genGetBind(expr *locals.GetBind) llvm.AnyValue {
	if expr.IsStatic() {
		return c.genIdentExpr(locals.NewIdentExpr(expr.Bind))
	}

	funcT, ok := asFuncType(expr.GetType())
	if !ok {
		panic(c.ice("bind type %s is not a func type", expr.GetType()))
	}
	params := stlslices.Map(funcT.GetParams(), func(i int, pt hir.Type) *hir.Param {
		return hir.NewParam(false, pt, fmt.Sprintf("p%d", i+1))
	})
	f := locals.NewFunc(funcT, params...)
	body := locals.NewBlock()
	f.Body = optional.Some(body)

	// self：标识符直接用其 Define；否则物化为局部 let（旧后端同做法，
	// 语义为「拷贝一份接收者」，方法体内对 self 的写入不影响原值）
	self := expr.From
	var selfIdent hir.Ident
	if identExpr, ok := self.(*locals.IdentExpr); ok {
		selfIdent = identExpr.Define
	} else {
		let := &locals.Let{
			Type:  self.GetType(),
			Value: optional.Some(expr.From),
		}
		c.genLocalLet(let)
		selfIdent = let
	}
	self = locals.NewIdentExpr(selfIdent)
	f.CaptureVariables = append(f.CaptureVariables, selfIdent)

	if refT, ok := asRefType(bindFirstParam(expr.Bind)); ok {
		self = locals.NewGetRef(refT.Mutable(), self)
	}
	args := []locals.Expr{self}
	for _, p := range params {
		args = append(args, locals.NewIdentExpr(p))
	}
	call := locals.NewCall(locals.NewIdentExpr(expr.Bind), args...)

	if _, isUnit := types.GetUnderlying(funcT.GetReturn()).(types.UnitType); isUnit {
		body.Stmts = append(body.Stmts, call, locals.NewReturn())
	} else {
		body.Stmts = append(body.Stmts, locals.NewReturn(call))
	}
	return c.genClosureValue(f)
}

// bindFirstParam 绑定函数的第一个参数类型（self）
func bindFirstParam(bind *locals.Let) hir.Type {
	funcT, ok := asFuncType(bind.GetType())
	if !ok || len(funcT.GetParams()) == 0 {
		panic(fmt.Errorf("codegen: bind %s has type %s, not a func type with self param", bind.Name, bind.GetType()))
	}
	return funcT.GetParams()[0]
}

// asRefType 引用类型断言（解包自定义类型别名）
func asRefType(t hir.Type) (types.RefType, bool) {
	rt, ok := types.GetUnderlying(t).(types.RefType)
	return rt, ok
}
