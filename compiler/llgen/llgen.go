package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
)

// LocalIdent 局部变量
type LocalIdent struct {
	Addr llvm.Value[llvm.PtrT] // 变量地址

	Func llvm.Value[llvm.FnT]  // 静态函数，可直接调用
	Ctx  llvm.Value[llvm.PtrT] // 闭包上下文，Func是闭包函数时有效
}

// CaptureVar 闭包捕获的变量
type CaptureVar struct {
	Ctx   llvm.Value[llvm.PtrT] // 闭包上下文地址
	Type  llvm.StructType       // 闭包上下文类型
	Index uint32                // 变量在上下文中的下标
}

type LLvmGenerator struct {
	pkg *globals.Package

	model   *ir.Module
	builder *ir.Builder

	ctx *Context

	currentFunc *locals.Func // 当前正在生成的函数
	llvmFunc    ir.Function  // 当前正在生成的LLVM函数

	locals         map[hir.Ident]*LocalIdent
	captureVarsMap map[tuple.Tuple2[*locals.Func, hir.Ident]]*CaptureVar
	eqFuncs        map[string]ir.Function

	funcCount int
	eqCount   int
	strCount  int
}

func New(ctx *Context, pkg *globals.Package) *LLvmGenerator {
	return &LLvmGenerator{
		pkg: pkg,

		model:   ir.NewModule(ctx.llvmCtx, pkg.Name),
		builder: ir.NewBuilder(ctx.llvmCtx),

		ctx: ctx,

		locals:         make(map[hir.Ident]*LocalIdent),
		captureVarsMap: make(map[tuple.Tuple2[*locals.Func, hir.Ident]]*CaptureVar),
		eqFuncs:        make(map[string]ir.Function),
	}
}

func (lg *LLvmGenerator) Generate() *ir.Module {
	for _, g := range lg.pkg.Globals {
		lg.genTypeDecl(g)
	}
	for _, g := range lg.pkg.Globals {
		lg.genTypeDef(g)
	}
	for _, g := range lg.pkg.Globals {
		lg.genGlobalDecl(g)
	}
	for _, g := range lg.pkg.Globals {
		lg.genGlobalValue(g)
	}

	if lg.pkg.Name == "main" {
		lg.genMainFunc()
	}

	stlerr.Must(lg.model.Verify())
	return lg.model
}

func (lg *LLvmGenerator) Close() error {
	err := stlerr.ErrorWrap(lg.builder.Close())
	if closeErr := lg.model.Close(); err == nil {
		err = stlerr.ErrorWrap(closeErr)
	}
	return err
}

// Model 返回正在生成的LLVM模块
func (lg *LLvmGenerator) Model() *ir.Module {
	return lg.model
}

// beginFunc 创建入口块并切换当前函数
func (lg *LLvmGenerator) beginFunc(fn ir.Function, hirFunc *locals.Func) ir.Block {
	entry := fn.NewBlock("entry")
	lg.llvmFunc = fn
	lg.currentFunc = hirFunc
	lg.builder.MoveToEnd(entry)
	return entry
}

// genFunctionBody 生成函数体，自动恢复构建器插入点
func (lg *LLvmGenerator) genFunctionBody(fn ir.Function, hirFunc *locals.Func, generate func()) {
	point := lg.builder.SaveInsertPoint()
	prevLLvmFunc := lg.llvmFunc
	prevHirFunc := lg.currentFunc

	lg.beginFunc(fn, hirFunc)
	generate()

	lg.llvmFunc = prevLLvmFunc
	lg.currentFunc = prevHirFunc
	lg.builder.RestoreInsertPoint(point)
}

// ensureBlock 当前基本块已终结时创建新的死代码块
func (lg *LLvmGenerator) ensureBlock() {
	cur, ok := lg.builder.CurrentBlock()
	if ok && !cur.IsTerminating() {
		return
	}
	lg.builder.MoveToEnd(lg.llvmFunc.NewBlock(""))
}

// ensureTerminator 基本块未终结时补上return
func (lg *LLvmGenerator) ensureTerminator(retType hir.Type) {
	cur, ok := lg.builder.CurrentBlock()
	if !ok || cur.IsTerminating() {
		return
	}
	if isUnitType(retType) {
		lg.builder.RetVoid()
	} else {
		lg.builder.Ret(lg.ctx.llvmCtx.ConstZero(lg.genType(retType).DynType()))
	}
}

// alloca 在函数入口块分配栈空间
func (lg *LLvmGenerator) alloca(t llvm.AnyType, name string) llvm.Value[llvm.PtrT] {
	point := lg.builder.SaveInsertPoint()
	entry, _ := lg.llvmFunc.EntryBlock()
	lg.builder.MoveToEnd(entry)
	if term, ok := entry.Terminator(); ok {
		lg.builder.MoveBefore(term)
	}
	a := lg.builder.Alloca(t, name)
	lg.builder.RestoreInsertPoint(point)
	return a.Value
}

func (lg *LLvmGenerator) uniqueFuncName() string {
	lg.funcCount++
	return stableName(lg.pkg, fmt.Sprintf(".func.%d", lg.funcCount))
}

func (lg *LLvmGenerator) uniqueStrName() string {
	lg.strCount++
	return stableName(lg.pkg, fmt.Sprintf(".str.%d", lg.strCount))
}

func (lg *LLvmGenerator) uniqueEqName() string {
	lg.eqCount++
	return stableName(lg.pkg, fmt.Sprintf(".eq.%d", lg.eqCount))
}

// load 从地址加载值
func (lg *LLvmGenerator) load(addr llvm.Value[llvm.PtrT], t hir.Type) llvm.AnyValue {
	return lg.builder.Load(addr, lg.genType(t).DynType(), "")
}

// store 将值写入地址
func (lg *LLvmGenerator) store(value llvm.AnyValue, addr llvm.Value[llvm.PtrT]) {
	lg.builder.Store(value, addr)
}

// buildFatFuncValue 构造函数胖值
func (lg *LLvmGenerator) buildFatFuncValue(fn llvm.Value[llvm.FnT], ctxv llvm.Value[llvm.PtrT]) llvm.AnyValue {
	llctx := lg.ctx.llvmCtx
	nullPtr := llctx.ConstNull(lg.ctx.ptrType)
	if ctxv.IsNil() {
		pair := llctx.ConstStruct(false, fn, nullPtr)
		return llctx.ConstStruct(false, pair, nullPtr)
	}
	if ctxv.IsConstant() {
		pair := llctx.ConstStruct(false, nullPtr, fn)
		return llctx.ConstStruct(false, pair, ctxv)
	}
	// 上下文非常量：用指令构造
	v := llctx.ConstZero(lg.ctx.fatFuncType.DynType())
	v = lg.builder.InsertValue(v, nullPtr, []uint32{0, 0}, "")
	v = lg.builder.InsertValue(v, fn, []uint32{0, 1}, "")
	v = lg.builder.InsertValue(v, ctxv, []uint32{1}, "")
	return v
}
