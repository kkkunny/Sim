package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (lg *LLvmGenerator) genLocal(local locals.Local) {
	switch local := local.(type) {
	case *locals.Block:
		lg.genBlockStmts(local)
	case *locals.Return:
		lg.genReturn(local)
	case *locals.Let:
		lg.genLocalLet(local)
	case locals.Expr:
		lg.genExpr(local)
	case *locals.If:
		lg.genIf(local, lg.newBlock("if.end"))
	case *locals.While:
		lg.genWhile(local)
	case *locals.For:
		lg.genFor(local)
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) newBlock(name string) ir.Block {
	return lg.llvmFunc.NewBlock(name)
}

func (lg *LLvmGenerator) genBlockStmts(b *locals.Block) {
	for _, s := range b.Stmts {
		lg.ensureBlock()
		lg.genLocal(s)
	}
}

// genNativeFuncParams 生成函数参数，offset为参数在LLVM函数中的起始下标（闭包的ctx参数占0）
func (lg *LLvmGenerator) genNativeFuncParams(fn ir.Function, params []*hir.Param, offset uint) {
	for i, p := range params {
		pn := fmt.Sprintf("_p%d", i+1)
		param := fn.Param(uint(i) + offset)
		param.SetName(pn)
		a := lg.alloca(lg.genType(p.Type), pn)
		lg.store(param, a)
		lg.locals[p] = &LocalIdent{Addr: a}
	}
}

func (lg *LLvmGenerator) genReturn(local *locals.Return) {
	v, ok := local.Value.Value()
	if !ok || isUnitType(v.GetType()) {
		if ok {
			lg.genExpr(v)
		}
		lg.builder.RetVoid()
		return
	}
	lg.builder.Ret(lg.genExpr(v))
}

func (lg *LLvmGenerator) genLocalLet(local *locals.Let) *LocalIdent {
	typ := lg.genType(local.GetType())
	value := local.Value.MustValue()
	if f, ok := value.(*locals.Func); ok && !local.Mut {
		fat, fn, ctxv := lg.genFuncValue(f)
		a := lg.alloca(typ, "")
		lg.store(fat, a)
		li := &LocalIdent{Addr: a, Func: fn, Ctx: ctxv}
		lg.locals[local] = li
		return li
	}

	v := lg.genExpr(value)
	a := lg.alloca(typ, "")
	lg.store(v, a)
	li := &LocalIdent{Addr: a}
	lg.locals[local] = li
	return li
}

func (lg *LLvmGenerator) genIf(l *locals.If, endBlock ir.Block) {
	cond := lg.genExpr(l.Condition)

	thenBlock := lg.newBlock("if.then")
	var elseBlock ir.Block
	if l.Else.IsSome() {
		elseBlock = lg.newBlock("if.else")
	} else {
		elseBlock = endBlock
	}
	lg.builder.CondBr(cond.Dyn().MustAs[llvm.IntT](), thenBlock, elseBlock)

	lg.builder.MoveToEnd(thenBlock)
	lg.genBlockStmts(l.Body)
	if cur, ok := lg.builder.CurrentBlock(); ok && !cur.IsTerminating() {
		lg.builder.Br(endBlock)
	}

	if next, ok := l.Else.Value(); ok {
		lg.builder.MoveToEnd(elseBlock)
		if elseif, ok := next.TryLeft(); ok {
			lg.genIf(elseif, endBlock)
		} else {
			lg.genBlockStmts(next.Right())
			if cur, ok := lg.builder.CurrentBlock(); ok && !cur.IsTerminating() {
				lg.builder.Br(endBlock)
			}
		}
	}
	lg.builder.MoveToEnd(endBlock)
}

func (lg *LLvmGenerator) genWhile(l *locals.While) {
	condBlock := lg.newBlock("while.cond")
	bodyBlock := lg.newBlock("while.body")
	endBlock := lg.newBlock("while.end")

	lg.builder.Br(condBlock)
	lg.builder.MoveToEnd(condBlock)
	lg.builder.CondBr(lg.genExpr(l.Condition).Dyn().MustAs[llvm.IntT](), bodyBlock, endBlock)

	lg.builder.MoveToEnd(bodyBlock)
	lg.genBlockStmts(l.Body)
	if cur, ok := lg.builder.CurrentBlock(); ok && !cur.IsTerminating() {
		lg.builder.Br(condBlock)
	}
	lg.builder.MoveToEnd(endBlock)
}

func (lg *LLvmGenerator) genFor(l *locals.For) {
	llctx := lg.ctx.llvmCtx
	i64 := llctx.Int(64)

	init := lg.alloca(i64, "")
	lg.store(i64.Const(0), init)

	rangePtr := lg.genExprAddr(l.Range)
	at := l.Range.GetType().(types.ArrayType)
	rangeT := lg.genType(l.Range.GetType())
	elemT := lg.genType(at.GetElem())
	varPtr := lg.alloca(elemT, "")

	condBlock := lg.newBlock("for.cond")
	bodyBlock := lg.newBlock("for.body")
	stepBlock := lg.newBlock("for.step")
	endBlock := lg.newBlock("for.end")

	lg.builder.Br(condBlock)

	lg.builder.MoveToEnd(condBlock)
	idx := lg.builder.Load(init, i64.DynType(), "").Dyn().MustAs[llvm.IntT]()
	cmp := lg.builder.ICmp(llvm.IntSLT, idx, llctx.ConstInt(i64, typeSize(at.GetSize())), "")
	lg.builder.CondBr(cmp, bodyBlock, endBlock)

	lg.builder.MoveToEnd(bodyBlock)
	elemPtr := lg.builder.InBoundsGEP(rangeT, rangePtr, []llvm.ValueRef[llvm.IntT]{
		lg.ctx.i32Type.Const(0), lg.ctx.i32Type.Const(0), idx,
	}, "")
	lg.store(lg.builder.Load(elemPtr, elemT.DynType(), ""), varPtr)
	lg.locals[l.Var] = &LocalIdent{Addr: varPtr}
	lg.genBlockStmts(l.Body)
	if cur, ok := lg.builder.CurrentBlock(); ok && !cur.IsTerminating() {
		lg.builder.Br(stepBlock)
	}

	lg.builder.MoveToEnd(stepBlock)
	lg.store(lg.builder.Add(idx, i64.Const(1), ""), init)
	lg.builder.Br(condBlock)

	lg.builder.MoveToEnd(endBlock)
}
