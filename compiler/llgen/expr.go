package llgen

import (
	"fmt"
	"math/big"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (lg *LLvmGenerator) genExpr(expr locals.Expr) llvm.AnyValue {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		return lg.genIdentExpr(expr)
	case *locals.Integer:
		return lg.genInteger(expr.Value, expr.GetType())
	case *locals.Float:
		v, _ := expr.Value.Float64()
		return llvm.MustFloatType(lg.genType(expr.GetType())).Const(v)
	case *locals.Boolean:
		return lg.ctx.llvmCtx.ConstBool(expr.Value)
	case *locals.String:
		return lg.genString(expr.Value)
	case locals.Unary:
		return lg.genUnary(expr)
	case *locals.Binary:
		return lg.genBinary(expr)
	case *locals.Func:
		v, _, _ := lg.genFuncValue(expr)
		return v
	case *locals.Call:
		return lg.genCall(expr)
	case *locals.Tuple:
		return lg.genTuple(expr)
	case *locals.TupleIndex:
		return lg.genTupleIndex(expr)
	case *locals.Array:
		return lg.genArray(expr)
	case *locals.ArrayIndex:
		return lg.genArrayIndex(expr)
	case locals.Covert:
		return lg.genCovert(expr)
	case *locals.Ternary:
		return lg.genTernary(expr)
	case *locals.Struct:
		return lg.genStruct(expr)
	case *locals.GetField:
		return lg.genGetField(expr)
	case *locals.GetBind:
		return lg.genGetBind(expr)
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) unitValue() llvm.AnyValue {
	return lg.ctx.llvmCtx.ConstZero(lg.ctx.emptyType)
}

func (lg *LLvmGenerator) genInteger(v *big.Int, t hir.Type) llvm.AnyValue {
	it := llvm.MustIntType(lg.genType(t))
	if stlval.Is[types.UintType](t) {
		return it.Const(v.Uint64())
	}
	return it.ConstS(v.Int64())
}

func (lg *LLvmGenerator) genString(s string) llvm.AnyValue {
	llctx := lg.ctx.llvmCtx
	c := llctx.ConstString(s, true)
	g := lg.model.NewGlobal(lg.uniqueStrName(), c.Type())
	g.SetInitializer(c)
	g.SetLinkage(llvm.LinkagePrivate)
	g.SetUnnamedAddr(llvm.UnnamedAddrGlobal)
	data := llctx.ConstGEP(c.Type(), g, true, lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0))
	return llctx.ConstStruct(false, data, lg.ctx.i64Type.Const(uint64(len(s))))
}

func (lg *LLvmGenerator) genIdentExpr(expr *locals.IdentExpr) llvm.AnyValue {
	if cv, ok := lg.captureVarsMap[tuple.Pack2(lg.currentFunc, expr.Define)]; ok {
		return lg.load(lg.captureAddr(cv), expr.GetType())
	}
	if li, ok := lg.locals[expr.Define]; ok {
		if !li.Func.IsNil() {
			return lg.buildFatFuncValue(li.Func, li.Ctx)
		}
		return lg.load(li.Addr, expr.GetType())
	}

	ident, ok := lg.ctx.idents[expr.Define]
	if !ok {
		panic("unreachable")
	}
	let, ok := expr.Define.(*locals.Let)
	if !ok {
		panic("unreachable")
	}
	if isFuncLet(let) {
		fn := lg.genGlobalFuncRef(let, ident)
		return lg.buildFatFuncValue(fn.Value, llvm.Value[llvm.PtrT]{})
	}
	g := lg.getOrDeclareVar(let, ident.Name, lg.genType(let.GetType()))
	return lg.load(g.Value, let.GetType())
}

// captureAddr 获取捕获变量的地址
func (lg *LLvmGenerator) captureAddr(cv *CaptureVar) llvm.Value[llvm.PtrT] {
	return lg.builder.InBoundsGEP(cv.Type, cv.Ctx, []llvm.ValueRef[llvm.IntT]{
		lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(uint64(cv.Index)),
	}, "")
}

// genGlobalFuncRef 获取全局函数的LLVM声明
func (lg *LLvmGenerator) genGlobalFuncRef(let *locals.Let, ident *Ident) ir.Function {
	if externalName, ok := let.ExternalName.Value(); ok && let.Value.IsNone() {
		return lg.getOrDeclareFunc(externalName, let.GetType().(types.FuncType))
	}
	return lg.getOrDeclareFunc(ident.Name, let.GetType().(types.FuncType))
}

// genIdentAddr 获取标识符地址
func (lg *LLvmGenerator) genIdentAddr(expr *locals.IdentExpr) llvm.Value[llvm.PtrT] {
	if cv, ok := lg.captureVarsMap[tuple.Pack2(lg.currentFunc, expr.Define)]; ok {
		return lg.captureAddr(cv)
	}

	if li, ok := lg.locals[expr.Define]; ok {
		if !li.Addr.IsNil() {
			return li.Addr
		}
		return li.Func.Dyn().MustAs[llvm.PtrT]()
	}

	ident, ok := lg.ctx.idents[expr.Define]
	if !ok {
		panic("unreachable")
	}
	let, ok := expr.Define.(*locals.Let)
	if !ok {
		panic("unreachable")
	}
	if isFuncLet(let) {
		return lg.genGlobalFuncRef(let, ident).Value.Dyn().MustAs[llvm.PtrT]()
	}
	g := lg.getOrDeclareVar(let, ident.Name, lg.genType(let.GetType()))
	return g.Value
}

func (lg *LLvmGenerator) genUnary(expr locals.Unary) llvm.AnyValue {
	switch expr := expr.(type) {
	case *locals.BitsReverse:
		return lg.builder.Not(lg.genExpr(expr.Target).Dyn().MustAs[llvm.IntT](), "")
	case *locals.BooleanReverse:
		v := lg.genExpr(expr.Target).Dyn().MustAs[llvm.IntT]()
		return lg.builder.Xor(v, lg.ctx.llvmCtx.ConstBool(true), "")
	case *locals.GetRef:
		return lg.genExprAddr(expr.Target)
	case *locals.DeRef:
		return lg.load(lg.genRefValue(expr.Target), expr.GetType())
	default:
		panic("unreachable")
	}
}

var AssignOp2BinaryOp = map[locals.BinaryOp]locals.BinaryOp{
	locals.BinaryOpEnum.AddAssign: locals.BinaryOpEnum.Add,
	locals.BinaryOpEnum.SubAssign: locals.BinaryOpEnum.Sub,
	locals.BinaryOpEnum.MulAssign: locals.BinaryOpEnum.Mul,
	locals.BinaryOpEnum.QuoAssign: locals.BinaryOpEnum.Quo,
	locals.BinaryOpEnum.RemAssign: locals.BinaryOpEnum.Rem,
	locals.BinaryOpEnum.AndAssign: locals.BinaryOpEnum.And,
	locals.BinaryOpEnum.OrAssign:  locals.BinaryOpEnum.Or,
	locals.BinaryOpEnum.XorAssign: locals.BinaryOpEnum.Xor,
	locals.BinaryOpEnum.ShlAssign: locals.BinaryOpEnum.Shl,
	locals.BinaryOpEnum.ShrAssign: locals.BinaryOpEnum.Shr,
}

func (lg *LLvmGenerator) genBinary(expr *locals.Binary) llvm.AnyValue {
	switch expr.Op {
	case locals.BinaryOpEnum.Assign:
		addr := lg.genExprAddr(expr.Left)
		lg.store(lg.genExpr(expr.Right), addr)
		return lg.unitValue()
	case locals.BinaryOpEnum.AddAssign, locals.BinaryOpEnum.SubAssign, locals.BinaryOpEnum.MulAssign,
		locals.BinaryOpEnum.QuoAssign, locals.BinaryOpEnum.RemAssign, locals.BinaryOpEnum.AndAssign,
		locals.BinaryOpEnum.OrAssign, locals.BinaryOpEnum.XorAssign, locals.BinaryOpEnum.ShlAssign,
		locals.BinaryOpEnum.ShrAssign:
		addr := lg.genExprAddr(expr.Left)
		left := lg.load(addr, expr.Left.GetType())
		right := lg.genExpr(expr.Right)
		lg.store(lg.genArith(AssignOp2BinaryOp[expr.Op], expr.Left.GetType(), left, right), addr)
		return lg.unitValue()
	case locals.BinaryOpEnum.Eq:
		return lg.genEqual(false, expr.Left.GetType(), lg.genExpr(expr.Left), lg.genExpr(expr.Right))
	case locals.BinaryOpEnum.Neq:
		return lg.genEqual(true, expr.Left.GetType(), lg.genExpr(expr.Left), lg.genExpr(expr.Right))
	case locals.BinaryOpEnum.LogicAnd, locals.BinaryOpEnum.LogicOr:
		return lg.genLogic(expr)
	default:
		return lg.genArith(expr.Op, expr.Left.GetType(), lg.genExpr(expr.Left), lg.genExpr(expr.Right))
	}
}

func (lg *LLvmGenerator) genArith(op locals.BinaryOp, t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	switch {
	case stlval.Is[types.IntegerType](t):
		return lg.genIntArith(op, t, left.Dyn().MustAs[llvm.IntT](), right.Dyn().MustAs[llvm.IntT]())
	case stlval.Is[types.FloatType](t):
		return lg.genFloatArith(op, left.Dyn().MustAs[llvm.FloatT](), right.Dyn().MustAs[llvm.FloatT]())
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genIntArith(op locals.BinaryOp, t hir.Type, left, right llvm.Value[llvm.IntT]) llvm.AnyValue {
	unsigned := stlval.Is[types.UintType](t)
	switch op {
	case locals.BinaryOpEnum.Add:
		return lg.builder.Add(left, right, "")
	case locals.BinaryOpEnum.Sub:
		return lg.builder.Sub(left, right, "")
	case locals.BinaryOpEnum.Mul:
		return lg.builder.Mul(left, right, "")
	case locals.BinaryOpEnum.Quo:
		if unsigned {
			return lg.builder.UDiv(left, right, "")
		}
		return lg.builder.SDiv(left, right, "")
	case locals.BinaryOpEnum.Rem:
		if unsigned {
			return lg.builder.URem(left, right, "")
		}
		return lg.builder.SRem(left, right, "")
	case locals.BinaryOpEnum.And:
		return lg.builder.And(left, right, "")
	case locals.BinaryOpEnum.Or:
		return lg.builder.Or(left, right, "")
	case locals.BinaryOpEnum.Xor:
		return lg.builder.Xor(left, right, "")
	case locals.BinaryOpEnum.Shl:
		return lg.builder.Shl(left, right, "")
	case locals.BinaryOpEnum.Shr:
		if unsigned {
			return lg.builder.LShr(left, right, "")
		}
		return lg.builder.AShr(left, right, "")
	case locals.BinaryOpEnum.Lt:
		return lg.builder.ICmp(stlval.If(unsigned, llvm.IntULT, llvm.IntSLT), left, right, "")
	case locals.BinaryOpEnum.Lte:
		return lg.builder.ICmp(stlval.If(unsigned, llvm.IntULE, llvm.IntSLE), left, right, "")
	case locals.BinaryOpEnum.Gt:
		return lg.builder.ICmp(stlval.If(unsigned, llvm.IntUGT, llvm.IntSGT), left, right, "")
	case locals.BinaryOpEnum.Gte:
		return lg.builder.ICmp(stlval.If(unsigned, llvm.IntUGE, llvm.IntSGE), left, right, "")
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genFloatArith(op locals.BinaryOp, left, right llvm.Value[llvm.FloatT]) llvm.AnyValue {
	switch op {
	case locals.BinaryOpEnum.Add:
		return lg.builder.FAdd(left, right, "")
	case locals.BinaryOpEnum.Sub:
		return lg.builder.FSub(left, right, "")
	case locals.BinaryOpEnum.Mul:
		return lg.builder.FMul(left, right, "")
	case locals.BinaryOpEnum.Quo:
		return lg.builder.FDiv(left, right, "")
	case locals.BinaryOpEnum.Rem:
		return lg.builder.FRem(left, right, "")
	case locals.BinaryOpEnum.Lt:
		return lg.builder.FCmp(llvm.FloatOLT, left, right, "")
	case locals.BinaryOpEnum.Lte:
		return lg.builder.FCmp(llvm.FloatOLE, left, right, "")
	case locals.BinaryOpEnum.Gt:
		return lg.builder.FCmp(llvm.FloatOGT, left, right, "")
	case locals.BinaryOpEnum.Gte:
		return lg.builder.FCmp(llvm.FloatOGE, left, right, "")
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genLogic(expr *locals.Binary) llvm.AnyValue {
	llctx := lg.ctx.llvmCtx
	left := lg.genExpr(expr.Left)
	leftBlock, _ := lg.builder.CurrentBlock()
	rhsBlock := lg.newBlock("logic.rhs")
	endBlock := lg.newBlock("logic.end")
	if expr.Op == locals.BinaryOpEnum.LogicAnd {
		lg.builder.CondBr(left.Dyn().MustAs[llvm.IntT](), rhsBlock, endBlock)
	} else {
		lg.builder.CondBr(left.Dyn().MustAs[llvm.IntT](), endBlock, rhsBlock)
	}

	lg.builder.MoveToEnd(rhsBlock)
	right := lg.genExpr(expr.Right)
	rhsEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(endBlock)
	phi := lg.builder.PHI(llctx.Bool(), "")
	if expr.Op == locals.BinaryOpEnum.LogicAnd {
		phi.AddIncoming(
			ir.Incoming[llvm.IntT]{Value: right.Dyn().MustAs[llvm.IntT](), Block: rhsEndBlock},
			ir.Incoming[llvm.IntT]{Value: llctx.ConstBool(false), Block: leftBlock},
		)
	} else {
		phi.AddIncoming(
			ir.Incoming[llvm.IntT]{Value: right.Dyn().MustAs[llvm.IntT](), Block: rhsEndBlock},
			ir.Incoming[llvm.IntT]{Value: llctx.ConstBool(true), Block: leftBlock},
		)
	}
	return phi
}

func (lg *LLvmGenerator) genCall(expr *locals.Call) llvm.AnyValue {
	ft := expr.Func.GetType().(types.FuncType)
	args := stlslices.Map(expr.Args, func(_ int, arg locals.Expr) llvm.AnyValue {
		return lg.genExpr(arg)
	})

	if fn, ctxv, ok := lg.tryGenStaticFunc(expr.Func); ok {
		if !ctxv.IsNil() {
			args = append([]llvm.AnyValue{ctxv}, args...)
		}
		res := lg.builder.Call[llvm.DynT](fn, args, "")
		if isUnitType(ft.GetReturn()) {
			return lg.unitValue()
		}
		return res
	}

	fv := lg.genExpr(expr.Func)
	return lg.genDynamicCall(fv, args, ft)
}

// tryGenStaticFunc 尝试获取可直接调用的静态函数
func (lg *LLvmGenerator) tryGenStaticFunc(expr locals.Expr) (llvm.Value[llvm.FnT], llvm.Value[llvm.PtrT], bool) {
	var zeroFn llvm.Value[llvm.FnT]
	var zeroPtr llvm.Value[llvm.PtrT]
	switch expr := expr.(type) {
	case *locals.Func:
		_, fn, ctxv := lg.genFuncValue(expr)
		return fn, ctxv, true
	case *locals.IdentExpr:
		if _, ok := lg.captureVarsMap[tuple.Pack2(lg.currentFunc, expr.Define)]; ok {
			return zeroFn, zeroPtr, false
		}
		let, ok := expr.Define.(*locals.Let)
		if !ok {
			return zeroFn, zeroPtr, false
		}
		if li, ok := lg.locals[let]; ok && !let.Mut && !li.Func.IsNil() {
			return li.Func, li.Ctx, true
		}
		if !isFuncLet(let) {
			return zeroFn, zeroPtr, false
		}
		ident, ok := lg.ctx.idents[let]
		if !ok {
			return zeroFn, zeroPtr, false
		}
		return lg.genGlobalFuncRef(let, ident).Value, zeroPtr, true
	default:
		return zeroFn, zeroPtr, false
	}
}

func (lg *LLvmGenerator) genDynamicCall(fv llvm.AnyValue, args []llvm.AnyValue, ft types.FuncType) llvm.AnyValue {
	llctx := lg.ctx.llvmCtx
	f := lg.extractPath(fv, []uint32{0, 0}).Dyn().MustAs[llvm.PtrT]()
	c := lg.extractPath(fv, []uint32{0, 1}).Dyn().MustAs[llvm.PtrT]()
	ctxv := lg.extractPath(fv, []uint32{1}).Dyn().MustAs[llvm.PtrT]()
	null := llctx.ConstNull(lg.ctx.ptrType)

	pureBlock := lg.newBlock("call.f")
	closureBlock := lg.newBlock("call.c")
	endBlock := lg.newBlock("call.end")
	lg.builder.CondBr(lg.builder.ICmp(llvm.IntEQ, ctxv, null, ""), pureBlock, closureBlock)

	retT := lg.genFuncReturnType(ft.GetReturn())
	params := stlslices.Map(ft.GetParams(), func(_ int, p hir.Type) llvm.AnyType {
		return lg.genType(p)
	})
	fSig := llctx.Fn(retT, params, false)
	cSig := llctx.Fn(retT, append([]llvm.AnyType{lg.ctx.ptrType}, params...), false)

	lg.builder.MoveToEnd(pureBlock)
	r1 := lg.builder.CallIndirect[llvm.DynT](f, fSig, args, "")
	pureEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(closureBlock)
	r2 := lg.builder.CallIndirect[llvm.DynT](c, cSig, append([]llvm.AnyValue{ctxv}, args...), "")
	closureEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(endBlock)
	if isUnitType(ft.GetReturn()) {
		return lg.unitValue()
	}
	phi := lg.builder.PHI(lg.genType(ft.GetReturn()).DynType(), "")
	phi.AddIncoming(
		ir.Incoming[llvm.DynT]{Value: r1.Dyn(), Block: pureEndBlock},
		ir.Incoming[llvm.DynT]{Value: r2.Dyn(), Block: closureEndBlock},
	)
	return phi
}

func (lg *LLvmGenerator) genFuncValue(expr *locals.Func) (llvm.AnyValue, llvm.Value[llvm.FnT], llvm.Value[llvm.PtrT]) {
	var zeroPtr llvm.Value[llvm.PtrT]
	if len(expr.CaptureVariables) == 0 {
		fn := lg.genNativeFuncDecl(expr)
		fn.SetLinkage(llvm.LinkageInternal)
		if b, ok := expr.Body.Value(); ok {
			lg.genFunctionBody(fn, expr, func() {
				lg.genNativeFuncParams(fn, expr.Params, 0)
				lg.genBlockStmts(b)
				lg.ensureTerminator(expr.Type.GetReturn())
			})
		}
		return lg.buildFatFuncValue(fn.Value, zeroPtr), fn.Value, zeroPtr
	}

	ctxT := lg.genClosureCtxType(expr.CaptureVariables)
	fn := lg.genNativeClosureDecl(expr, ctxT)
	ctxPtr := lg.buildClosureCtx(expr, ctxT)
	return lg.buildFatFuncValue(fn.Value, ctxPtr), fn.Value, ctxPtr
}

func (lg *LLvmGenerator) genNativeFuncDecl(expr *locals.Func) ir.Function {
	ft := expr.GetType().(types.FuncType)
	return lg.model.NewFunction(lg.uniqueFuncName(), lg.genNativeFuncType(ft))
}

func (lg *LLvmGenerator) genClosureCtxType(captures []hir.Ident) llvm.StructType {
	fields := stlslices.Map(captures, func(_ int, cv hir.Ident) llvm.AnyType {
		return lg.genType(cv.GetType())
	})
	return lg.ctx.llvmCtx.Struct(fields, false)
}

func (lg *LLvmGenerator) genNativeClosureDecl(expr *locals.Func, ctxT llvm.StructType) ir.Function {
	ft := expr.GetType().(types.FuncType)
	params := append([]llvm.AnyType{lg.ctx.ptrType}, stlslices.Map(expr.Params, func(_ int, p *hir.Param) llvm.AnyType {
		return lg.genType(p.Type)
	})...)
	fn := lg.model.NewFunction(lg.uniqueFuncName(), lg.ctx.llvmCtx.Fn(lg.genFuncReturnType(ft.GetReturn()), params, false))
	fn.SetLinkage(llvm.LinkageInternal)

	if b, ok := expr.Body.Value(); ok {
		lg.genFunctionBody(fn, expr, func() {
			ctxParam := fn.Param(0)
			lg.genNativeFuncParams(fn, expr.Params, 1)
			ctxAddr := lg.alloca(ctxT, "_ctx")
			lg.store(lg.builder.Load(ctxParam.Dyn().MustAs[llvm.PtrT](), ctxT.DynType(), ""), ctxAddr)
			for i, cv := range expr.CaptureVariables {
				lg.captureVarsMap[tuple.Pack2(expr, cv)] = &CaptureVar{
					Ctx:   ctxAddr,
					Type:  ctxT,
					Index: uint32(i),
				}
			}
			lg.genBlockStmts(b)
			lg.ensureTerminator(ft.GetReturn())
		})
	}
	return fn
}

// buildClosureCtx 在闭包创建处构造上下文
func (lg *LLvmGenerator) buildClosureCtx(expr *locals.Func, ctxT llvm.StructType) llvm.Value[llvm.PtrT] {
	ctxAddr := lg.alloca(ctxT, "")
	for i, cv := range expr.CaptureVariables {
		fieldPtr := lg.builder.InBoundsGEP(ctxT, ctxAddr, []llvm.ValueRef[llvm.IntT]{
			lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(uint64(i)),
		}, "")
		lg.store(lg.genExpr(locals.NewIdentExpr(cv)), fieldPtr)
	}
	return ctxAddr
}

func (lg *LLvmGenerator) genTuple(expr *locals.Tuple) llvm.AnyValue {
	t := lg.genType(expr.GetType())
	v := lg.ctx.llvmCtx.ConstZero(t.DynType())
	for i, e := range expr.Elems {
		v = lg.builder.InsertValue(v, lg.genExpr(e), []uint32{uint32(i)}, "")
	}
	return v
}

func (lg *LLvmGenerator) genTupleIndex(expr *locals.TupleIndex) llvm.AnyValue {
	if !expr.From.Temporary() {
		return lg.load(lg.genTupleIndexAddr(expr), expr.GetType())
	}
	return lg.extractPath(lg.genExpr(expr.From), []uint32{uint32(expr.Index.Uint64())})
}

func (lg *LLvmGenerator) genTupleIndexAddr(expr *locals.TupleIndex) llvm.Value[llvm.PtrT] {
	addr := lg.genExprAddr(expr.From)
	st := lg.genType(expr.From.GetType())
	return lg.builder.InBoundsGEP(st, addr, []llvm.ValueRef[llvm.IntT]{
		lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(expr.Index.Uint64()),
	}, "")
}

func (lg *LLvmGenerator) genArray(expr *locals.Array) llvm.AnyValue {
	t := lg.genType(expr.GetType())
	v := lg.ctx.llvmCtx.ConstZero(t.DynType())
	for i, e := range expr.Elems {
		v = lg.builder.InsertValue(v, lg.genExpr(e), []uint32{0, uint32(i)}, "")
	}
	return v
}

func (lg *LLvmGenerator) genArrayIndex(expr *locals.ArrayIndex) llvm.AnyValue {
	return lg.load(lg.genArrayIndexAddr(expr), expr.GetType())
}

func (lg *LLvmGenerator) genArrayIndexAddr(expr *locals.ArrayIndex) llvm.Value[llvm.PtrT] {
	addr := lg.genExprAddr(expr.From)
	index := lg.genArrayIndexValue(expr.Index)
	at := lg.genType(expr.From.GetType())
	return lg.builder.InBoundsGEP(at, addr, []llvm.ValueRef[llvm.IntT]{
		lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0), index,
	}, "")
}

func (lg *LLvmGenerator) genArrayIndexValue(index locals.Expr) llvm.Value[llvm.IntT] {
	v := lg.genExpr(index).Dyn().MustAs[llvm.IntT]()
	if llvm.MustIntType(lg.genType(index.GetType())).Bits() == 64 {
		return v
	}
	return lg.builder.IntCast(v, lg.ctx.i64Type, !stlval.Is[types.UintType](index.GetType()), "")
}

func (lg *LLvmGenerator) genCovert(expr locals.Covert) llvm.AnyValue {
	v := lg.genExpr(expr.GetFrom())
	switch expr := expr.(type) {
	case *locals.Union:
		uT := lg.genType(expr.GetType())
		a := lg.alloca(uT, "")
		tagPtr := lg.builder.InBoundsGEP(uT, a, []llvm.ValueRef[llvm.IntT]{
			lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0),
		}, "")
		lg.store(lg.ctx.i8Type.Const(uint64(expr.Index)), tagPtr)
		valuePtr := lg.builder.InBoundsGEP(uT, a, []llvm.ValueRef[llvm.IntT]{
			lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(1),
		}, "")
		lg.store(v, valuePtr)
		return lg.load(a, expr.GetType())
	case *locals.NumberCovert:
		return lg.genNumberCovert(expr.GetFrom().GetType(), expr.GetType(), v)
	case *locals.TypedefCovert:
		fromT := lg.genType(expr.GetFrom().GetType())
		toT := lg.genType(expr.GetType())
		if fromT.Equal(toT) {
			return v
		}
		a := lg.alloca(toT, "")
		lg.store(v, a)
		return lg.load(a, expr.GetType())
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genNumberCovert(fromT, toT hir.Type, v llvm.AnyValue) llvm.AnyValue {
	fromU := types.GetUnderlying(fromT)
	toU := types.GetUnderlying(toT)
	fromInt := stlval.Is[types.IntegerType](fromU)
	toInt := stlval.Is[types.IntegerType](toU)
	fromFloat := stlval.Is[types.FloatType](fromU)
	toFloat := stlval.Is[types.FloatType](toU)
	switch {
	case fromInt && toInt:
		to := llvm.MustIntType(lg.genType(toT))
		return lg.builder.IntCast(v.Dyn().MustAs[llvm.IntT](), to, !stlval.Is[types.UintType](fromU), "")
	case fromInt && toFloat:
		to := llvm.MustFloatType(lg.genType(toT))
		if stlval.Is[types.UintType](fromU) {
			return lg.builder.UIToFP(v.Dyn().MustAs[llvm.IntT](), to, "")
		}
		return lg.builder.SIToFP(v.Dyn().MustAs[llvm.IntT](), to, "")
	case fromFloat && toInt:
		to := llvm.MustIntType(lg.genType(toT))
		if stlval.Is[types.UintType](toU) {
			return lg.builder.FPToUI(v.Dyn().MustAs[llvm.FloatT](), to, "")
		}
		return lg.builder.FPToSI(v.Dyn().MustAs[llvm.FloatT](), to, "")
	case fromFloat && toFloat:
		to := llvm.MustFloatType(lg.genType(toT))
		if llvm.MustFloatType(lg.genType(fromT)).Kind() == llvm.FloatSingle {
			return lg.builder.FPExt(v.Dyn().MustAs[llvm.FloatT](), to, "")
		}
		return lg.builder.FPTrunc(v.Dyn().MustAs[llvm.FloatT](), to, "")
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genTernary(expr *locals.Ternary) llvm.AnyValue {
	cond := lg.genExpr(expr.Condition)
	thenBlock := lg.newBlock("ternary.then")
	elseBlock := lg.newBlock("ternary.else")
	endBlock := lg.newBlock("ternary.end")
	lg.builder.CondBr(cond.Dyn().MustAs[llvm.IntT](), thenBlock, elseBlock)

	lg.builder.MoveToEnd(thenBlock)
	thenValue := lg.genExpr(expr.TrueExpr)
	thenEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(elseBlock)
	elseValue := lg.genExpr(expr.FalseExpr)
	elseEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(endBlock)
	if isUnitType(expr.GetType()) {
		return lg.unitValue()
	}
	phi := lg.builder.PHI(lg.genType(expr.GetType()).DynType(), "")
	phi.AddIncoming(
		ir.Incoming[llvm.DynT]{Value: thenValue.Dyn(), Block: thenEndBlock},
		ir.Incoming[llvm.DynT]{Value: elseValue.Dyn(), Block: elseEndBlock},
	)
	return phi
}

func (lg *LLvmGenerator) genEqual(not bool, t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	res := lg.genEqualValue(t, left, right)
	if not {
		return lg.builder.Xor(res.Dyn().MustAs[llvm.IntT](), lg.ctx.llvmCtx.ConstBool(true), "")
	}
	return res
}

func (lg *LLvmGenerator) genEqualValue(t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	switch t := t.(type) {
	case types.CustomType:
		underlying := types.GetUnderlying(t)
		switch underlying.(type) {
		case types.RefType:
			return lg.builder.ICmp(llvm.IntEQ, lg.extractPath(left, []uint32{0}).Dyn().MustAs[llvm.PtrT](), lg.extractPath(right, []uint32{0}).Dyn().MustAs[llvm.PtrT](), "")
		case types.IntegerType, types.FloatType, types.BooleanType, types.StringType:
			return lg.genEqualValue(underlying, left, right)
		default:
			return lg.genEqualCall(t, left, right)
		}
	case types.IntegerType, types.BooleanType:
		return lg.builder.ICmp(llvm.IntEQ, left.Dyn().MustAs[llvm.IntT](), right.Dyn().MustAs[llvm.IntT](), "")
	case types.FloatType:
		return lg.builder.FCmp(llvm.FloatOEQ, left.Dyn().MustAs[llvm.FloatT](), right.Dyn().MustAs[llvm.FloatT](), "")
	case types.StringType:
		return lg.genStringEqual(left, right)
	case types.RefType:
		return lg.builder.ICmp(llvm.IntEQ, left.Dyn().MustAs[llvm.PtrT](), right.Dyn().MustAs[llvm.PtrT](), "")
	default:
		return lg.genEqualCall(t, left, right)
	}
}

func (lg *LLvmGenerator) genEqualCall(t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	fn := lg.genEqualFunc(t)
	return lg.builder.Call[llvm.DynT](fn, []llvm.AnyValue{left, right}, "")
}

func (lg *LLvmGenerator) genStringEqual(left, right llvm.AnyValue) llvm.AnyValue {
	llctx := lg.ctx.llvmCtx
	lData, rData := lg.extractPath(left, []uint32{0}), lg.extractPath(right, []uint32{0})
	lLen := lg.extractPath(left, []uint32{1}).Dyn().MustAs[llvm.IntT]()
	rLen := lg.extractPath(right, []uint32{1}).Dyn().MustAs[llvm.IntT]()
	lenEq := lg.builder.ICmp(llvm.IntEQ, lLen, rLen, "")

	memcmp := lg.getOrDeclareMemcmp()
	thenBlock := lg.newBlock("streq.then")
	elseBlock := lg.newBlock("streq.else")
	endBlock := lg.newBlock("streq.end")
	lg.builder.CondBr(lenEq, thenBlock, elseBlock)

	lg.builder.MoveToEnd(thenBlock)
	res := lg.builder.Call[llvm.DynT](memcmp, []llvm.AnyValue{lData, rData, lLen}, "")
	cmpZero := lg.builder.ICmp(llvm.IntEQ, res.Dyn().MustAs[llvm.IntT](), llctx.Int(32).Const(0), "").Dyn().MustAs[llvm.IntT]()
	thenEndBlock, _ := lg.builder.CurrentBlock()
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(elseBlock)
	lg.builder.Br(endBlock)

	lg.builder.MoveToEnd(endBlock)
	phi := lg.builder.PHI(llctx.Bool(), "")
	phi.AddIncoming(
		ir.Incoming[llvm.IntT]{Value: cmpZero, Block: thenEndBlock},
		ir.Incoming[llvm.IntT]{Value: llctx.ConstBool(false), Block: elseBlock},
	)
	return phi
}

func (lg *LLvmGenerator) getOrDeclareMemcmp() ir.Function {
	if fn, ok := lg.model.GetFunction("memcmp"); ok {
		return fn
	}
	llctx := lg.ctx.llvmCtx
	params := []llvm.AnyType{lg.ctx.ptrType, lg.ctx.ptrType, llctx.Int(64)}
	return lg.model.NewFunction("memcmp", llctx.Fn(llctx.Int(32), params, false))
}

func (lg *LLvmGenerator) genEqualFunc(t hir.Type) ir.Function {
	key := t.String()
	if fn, ok := lg.eqFuncs[key]; ok {
		return fn
	}

	llctx := lg.ctx.llvmCtx
	tt := lg.genType(t)
	fn := lg.model.NewFunction(lg.uniqueEqName(), llctx.Fn(llctx.Bool(), []llvm.AnyType{tt, tt}, false))
	fn.SetLinkage(llvm.LinkageInternal)
	lg.eqFuncs[key] = fn

	lg.genFunctionBody(fn, nil, func() {
		lg.genEqualFuncBody(t, fn.Param(0), fn.Param(1))
		lg.ensureTerminator(types.Bool)
	})
	return fn
}

func (lg *LLvmGenerator) genEqualFuncBody(t hir.Type, x, y llvm.AnyValue) {
	switch t := types.GetUnderlying(t).(type) {
	case types.ArrayType:
		lg.genArrayEqualBody(t, x, y)
	case types.TupleType:
		lg.genFieldsEqualBody(t.GetElems(), x, y)
	case types.StructType:
		fields := stlslices.Map(t.GetFields(), func(_ int, f *types.StructField) hir.Type {
			return f.Type
		})
		lg.genFieldsEqualBody(fields, x, y)
	case types.UnionType:
		lg.genUnionEqualBody(t, x, y)
	case types.FuncType:
		lg.genFuncEqualBody(x, y)
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genArrayEqualBody(t types.ArrayType, x, y llvm.AnyValue) {
	llctx := lg.ctx.llvmCtx
	i64 := llctx.Int(64)
	at := lg.genType(t)
	elemT := lg.genType(t.GetElem())

	xAddr := lg.alloca(at, "")
	lg.store(x, xAddr)
	yAddr := lg.alloca(at, "")
	lg.store(y, yAddr)
	idxAddr := lg.alloca(i64, "")
	lg.store(i64.Const(0), idxAddr)

	condBlock := lg.newBlock("eq.cond")
	bodyBlock := lg.newBlock("eq.body")
	retBlock := lg.newBlock("eq.ret")
	contBlock := lg.newBlock("eq.cont")
	endBlock := lg.newBlock("eq.end")

	lg.builder.Br(condBlock)
	lg.builder.MoveToEnd(condBlock)
	idx := lg.builder.Load(idxAddr, i64.DynType(), "").Dyn().MustAs[llvm.IntT]()
	cmp := lg.builder.ICmp(llvm.IntSLT, idx, llctx.ConstInt(i64, typeSize(t.GetSize())), "")
	lg.builder.CondBr(cmp, bodyBlock, endBlock)

	lg.builder.MoveToEnd(bodyBlock)
	index := []llvm.ValueRef[llvm.IntT]{lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0), idx}
	lv := lg.builder.Load(lg.builder.InBoundsGEP(at, xAddr, index, ""), elemT.DynType(), "")
	rv := lg.builder.Load(lg.builder.InBoundsGEP(at, yAddr, index, ""), elemT.DynType(), "")
	lg.builder.CondBr(lg.genEqual(true, t.GetElem(), lv, rv).Dyn().MustAs[llvm.IntT](), retBlock, contBlock)

	lg.builder.MoveToEnd(retBlock)
	lg.builder.Ret(llctx.ConstBool(false))

	lg.builder.MoveToEnd(contBlock)
	lg.store(lg.builder.Add(idx, i64.Const(1), ""), idxAddr)
	lg.builder.Br(condBlock)

	lg.builder.MoveToEnd(endBlock)
	lg.builder.Ret(llctx.ConstBool(true))
}

func (lg *LLvmGenerator) genFieldsEqualBody(fields []hir.Type, x, y llvm.AnyValue) {
	llctx := lg.ctx.llvmCtx
	for i, field := range fields {
		lv := lg.extractPath(x, []uint32{uint32(i)})
		rv := lg.extractPath(y, []uint32{uint32(i)})
		retBlock := lg.newBlock("eq.ret")
		contBlock := lg.newBlock("eq.cont")
		lg.builder.CondBr(lg.genEqual(true, field, lv, rv).Dyn().MustAs[llvm.IntT](), retBlock, contBlock)
		lg.builder.MoveToEnd(retBlock)
		lg.builder.Ret(llctx.ConstBool(false))
		lg.builder.MoveToEnd(contBlock)
	}
	lg.builder.Ret(llctx.ConstBool(true))
}

func (lg *LLvmGenerator) genUnionEqualBody(t types.UnionType, x, y llvm.AnyValue) {
	llctx := lg.ctx.llvmCtx
	i8 := llctx.Int(8)
	uT := lg.genType(t)

	xAddr := lg.alloca(uT, "")
	lg.store(x, xAddr)
	yAddr := lg.alloca(uT, "")
	lg.store(y, yAddr)

	tagIndex := []llvm.ValueRef[llvm.IntT]{lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0)}
	valueIndex := []llvm.ValueRef[llvm.IntT]{lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(1)}
	xTag := lg.builder.Load(lg.builder.InBoundsGEP(uT, xAddr, tagIndex, ""), i8.DynType(), "").Dyn().MustAs[llvm.IntT]()
	yTag := lg.builder.Load(lg.builder.InBoundsGEP(uT, yAddr, tagIndex, ""), i8.DynType(), "").Dyn().MustAs[llvm.IntT]()

	diffBlock := lg.newBlock("union.diff")
	sameBlock := lg.newBlock("union.same")
	defaultBlock := lg.newBlock("union.default")
	lg.builder.CondBr(lg.builder.ICmp(llvm.IntNE, xTag, yTag, ""), diffBlock, sameBlock)

	lg.builder.MoveToEnd(diffBlock)
	lg.builder.Ret(llctx.ConstBool(false))

	lg.builder.MoveToEnd(sameBlock)
	sw := lg.builder.Switch(xTag, defaultBlock)
	elems := t.GetElems()
	caseBlocks := stlslices.Map(elems, func(i int, _ hir.Type) ir.Block {
		block := lg.newBlock(fmt.Sprintf("union.case%d", i))
		sw.AddCase(i8.Const(uint64(i)), block)
		return block
	})
	for i, elem := range elems {
		lg.builder.MoveToEnd(caseBlocks[i])
		elemT := lg.genType(elem)
		xv := lg.builder.Load(lg.builder.InBoundsGEP(uT, xAddr, valueIndex, ""), elemT.DynType(), "")
		yv := lg.builder.Load(lg.builder.InBoundsGEP(uT, yAddr, valueIndex, ""), elemT.DynType(), "")
		lg.builder.Ret(lg.genEqual(false, elem, xv, yv))
	}

	lg.builder.MoveToEnd(defaultBlock)
	lg.builder.Ret(llctx.ConstBool(true))
}

func (lg *LLvmGenerator) genFuncEqualBody(x, y llvm.AnyValue) {
	llctx := lg.ctx.llvmCtx
	null := llctx.ConstNull(lg.ctx.ptrType)
	ctxL := lg.extractPath(x, []uint32{1}).Dyn().MustAs[llvm.PtrT]()
	ctxR := lg.extractPath(y, []uint32{1}).Dyn().MustAs[llvm.PtrT]()
	fL := lg.extractPath(x, []uint32{0, 0}).Dyn().MustAs[llvm.PtrT]()
	cL := lg.extractPath(x, []uint32{0, 1}).Dyn().MustAs[llvm.PtrT]()
	fR := lg.extractPath(y, []uint32{0, 0}).Dyn().MustAs[llvm.PtrT]()
	cR := lg.extractPath(y, []uint32{0, 1}).Dyn().MustAs[llvm.PtrT]()

	sameCtx := lg.builder.ICmp(llvm.IntEQ, ctxL, ctxR, "")
	bothNull := lg.builder.ICmp(llvm.IntEQ, ctxL, null, "")
	fnEq := lg.builder.Select(bothNull, lg.builder.ICmp(llvm.IntEQ, fL, fR, ""), lg.builder.ICmp(llvm.IntEQ, cL, cR, ""), "")
	lg.builder.Ret(lg.builder.And(sameCtx, fnEq, ""))
}

func (lg *LLvmGenerator) genStruct(expr *locals.Struct) llvm.AnyValue {
	t := lg.genType(expr.GetType())
	v := lg.ctx.llvmCtx.ConstZero(t.DynType())
	for i, field := range expr.Type.GetFields() {
		index := []uint32{uint32(i)}
		e, ok := expr.Fields[field.Name]
		if !ok {
			v = lg.builder.InsertValue(v, lg.ctx.llvmCtx.ConstZero(lg.genType(field.Type).DynType()), index, "")
			continue
		}
		v = lg.builder.InsertValue(v, lg.genExpr(e), index, "")
	}
	return v
}

func (lg *LLvmGenerator) genGetField(expr *locals.GetField) llvm.AnyValue {
	if !expr.From.Temporary() {
		return lg.load(lg.genGetFieldAddr(expr), expr.GetType())
	}
	index := fieldIndex(expr.From.GetType(), expr.Name)
	return lg.extractPath(lg.genExpr(expr.From), []uint32{index})
}

func (lg *LLvmGenerator) genGetFieldAddr(expr *locals.GetField) llvm.Value[llvm.PtrT] {
	addr := lg.genExprAddr(expr.From)
	st := lg.genType(expr.From.GetType())
	index := fieldIndex(expr.From.GetType(), expr.Name)
	return lg.builder.InBoundsGEP(st, addr, []llvm.ValueRef[llvm.IntT]{
		lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(uint64(index)),
	}, "")
}

func fieldIndex(t hir.Type, name string) uint32 {
	st, ok := t.(types.StructType)
	if !ok {
		panic("unreachable")
	}
	for i, field := range st.GetFields() {
		if field.Name == name {
			return uint32(i)
		}
	}
	panic("unreachable")
}

func (lg *LLvmGenerator) genGetBind(expr *locals.GetBind) llvm.AnyValue {
	if expr.IsStatic() {
		return lg.genIdentExpr(locals.NewIdentExpr(expr.Bind))
	}

	ft := expr.GetType().(types.FuncType)
	params := stlslices.Map(ft.GetParams(), func(i int, paramType hir.Type) *hir.Param {
		return hir.NewParam(false, paramType, fmt.Sprintf("p%d", i+1))
	})
	f := locals.NewFunc(ft, params...)
	body := locals.NewBlock()
	f.Body = optional.Some(body)

	self := expr.From
	var selfIdent hir.Ident
	if identExpr, ok := self.(*locals.IdentExpr); ok {
		selfIdent = identExpr.Define
	} else {
		let := &locals.Let{
			Type:  self.GetType(),
			Value: optional.Some(expr.From),
		}
		lg.genLocalLet(let)
		selfIdent = let
	}
	self = locals.NewIdentExpr(selfIdent)
	f.CaptureVariables = append(f.CaptureVariables, selfIdent)

	if refT, ok := expr.Bind.GetType().(types.FuncType).GetParams()[0].(types.RefType); ok {
		self = locals.NewGetRef(refT.Mutable(), self)
	}
	args := []locals.Expr{self}
	for _, p := range params {
		args = append(args, locals.NewIdentExpr(p))
	}
	call := locals.NewCall(locals.NewIdentExpr(expr.Bind), args...)

	if !ft.GetReturn().Equal(types.Unit) {
		body.Stmts = append(body.Stmts, locals.NewReturn(call))
	} else {
		body.Stmts = append(body.Stmts, call, locals.NewReturn())
	}
	return lg.genExpr(f)
}

// genExprAddr 获取表达式地址
func (lg *LLvmGenerator) genExprAddr(expr locals.Expr) llvm.Value[llvm.PtrT] {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		return lg.genIdentAddr(expr)
	case *locals.DeRef:
		return lg.genRefValue(expr.Target)
	case *locals.GetField:
		return lg.genGetFieldAddr(expr)
	case *locals.TupleIndex:
		return lg.genTupleIndexAddr(expr)
	case *locals.ArrayIndex:
		return lg.genArrayIndexAddr(expr)
	default:
		v := lg.genExpr(expr)
		a := lg.alloca(lg.genType(expr.GetType()), "")
		lg.store(v, a)
		return a
	}
}

// genRefValue 获取引用表达式的指针
func (lg *LLvmGenerator) genRefValue(expr locals.Expr) llvm.Value[llvm.PtrT] {
	if stlval.Is[types.CustomType](expr.GetType()) {
		return lg.extractPath(lg.genExpr(expr), []uint32{0}).Dyn().MustAs[llvm.PtrT]()
	}
	return lg.genExpr(expr).Dyn().MustAs[llvm.PtrT]()
}

func (lg *LLvmGenerator) extractPath(value llvm.AnyValue, indices []uint32) llvm.AnyValue {
	return lg.builder.ExtractValue[llvm.DynT](value, indices, "")
}
