package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// CodeGenerator HIR → LLVM IR 代码生成器（每个包一个实例、一个 LLVM 模块）
type CodeGenerator struct {
	pkg     *globals.Package
	ctx     *Context
	module  *ir.Module
	builder *ir.Builder

	currentFunc ir.Function // 当前正在生成的函数
	terminated  bool        // 当前基本块是否已终结
	strCount    int         // 字符串字面量计数
	moduleID    int         // 模块标识（相等性辅助函数缓存按模块隔离）

	captureVars map[hir.Ident]llvm.Value[llvm.PtrT] // 当前包装函数：捕获变量 → ctx 字段地址（F7）
	closures    map[*locals.Func]*closureInfo       // 闭包字面量 → 内部函数/捕获布局（按字面量去重）
}

// New 创建某包的代码生成器
func New(ctx *Context, pkg *globals.Package) *CodeGenerator {
	module := ir.NewModule(ctx.LLVM(), pkg.Name)
	ctx.TargetMachine().ApplyTo(module)
	ctx.moduleCount++
	return &CodeGenerator{
		pkg:      pkg,
		ctx:      ctx,
		module:   module,
		builder:  ir.NewBuilder(ctx.LLVM()),
		moduleID: ctx.moduleCount,

		closures: make(map[*locals.Func]*closureInfo),
	}
}

// Module 返回当前包对应的 LLVM 模块
func (c *CodeGenerator) Module() *ir.Module {
	return c.module
}

// Close 释放当前生成器资源
func (c *CodeGenerator) Close() error {
	err1 := c.builder.Close()
	err2 := c.module.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Generate 生成包对应的 LLVM 模块
func (c *CodeGenerator) Generate() *ir.Module {
	for _, g := range c.pkg.Globals {
		c.genTypeDecl(g)
	}
	for _, g := range c.pkg.Globals {
		c.genTypeDef(g)
	}
	// 函数符号声明子遍：先登记全部函数符号，再生成函数体，支持同包前向引用/互递归
	c.genFuncDecls()
	// 全局变量符号声明子遍：支持函数体前向引用同包后置声明的全局变量
	c.genGlobalVarDecls()
	for _, g := range c.pkg.Globals {
		c.genGlobalValue(g)
	}
	if c.pkg.Name == "main" {
		c.genEntryWrapper()
	}

	if err := c.module.Verify(); err != nil {
		panic(fmt.Errorf("包 %s 生成 LLVM IR 校验失败: %w", c.pkg.Path, err))
	}
	return c.module
}

func (c *CodeGenerator) moveTo(block ir.Block) {
	c.builder.MoveToEnd(block)
	c.terminated = false
}

// ensureBlock 当前块已终结时，新建一个不可达的块继续发射（死代码）
func (c *CodeGenerator) ensureBlock() {
	if !c.terminated {
		return
	}
	c.moveTo(c.currentFunc.NewBlock(""))
}

// genFuncBody 生成函数体（F1）：登记参数存储 → 可选初始化（闭包 ctx 捕获）→ 逐语句生成，
// 最后按返回类型补 ret void/unreachable。函数体生成期间保存并恢复调用点的发射状态
// （闭包包装函数在表达式求值中途嵌套生成）。
//
// paramOffset 为 HIR 参数在 LLVM 形参列表中的起始下标：普通函数为 0，
// 闭包包装函数首参为 ctx ptr，偏移 1。
func (c *CodeGenerator) genFuncBody(decl ir.Function, expr *locals.Func, body *locals.Block, paramOffset int, initFn func()) {
	prevFunc, prevTerminated := c.currentFunc, c.terminated
	prevBlock, hadBlock := c.builder.CurrentBlock()
	defer func() {
		c.currentFunc = prevFunc
		c.terminated = prevTerminated
		if hadBlock {
			c.builder.MoveToEnd(prevBlock)
		}
	}()

	c.currentFunc = decl
	c.moveTo(decl.NewBlock("entry"))
	for i, p := range expr.Params {
		ptr := c.builder.Alloca(c.genType(p.GetType()), p.GetName())
		c.builder.Store(decl.Param(uint(i+paramOffset)), ptr)
		c.ctx.idents[p] = &Ident{Name: p.GetName(), Local: ptr.Value}
	}
	if initFn != nil {
		initFn()
	}
	for _, s := range body.Stmts {
		c.genLocal(s)
	}
	if !c.terminated {
		if _, ok := expr.Type.GetReturn().(types.UnitType); ok {
			c.builder.RetVoid()
		} else {
			c.builder.Unreachable()
		}
		c.terminated = true
	}
}

// genEntryWrapper 生成 C 入口 main（调用 sim_main）
func (c *CodeGenerator) genEntryWrapper() {
	i32 := c.ctx.LLVM().Int(32)
	fn := c.module.NewFunction("main", c.ctx.LLVM().Fn(i32, nil, false))
	c.currentFunc = fn
	c.moveTo(fn.NewBlock("entry"))
	if simMain, ok := c.module.GetFunction("sim_main"); ok {
		c.builder.Call[llvm.DynT](simMain, nil, "")
	}
	c.builder.Ret(i32.Const(0))
	c.terminated = true
}
