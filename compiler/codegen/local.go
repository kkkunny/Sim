package codegen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"

	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genLocal(local locals.Local) {
	switch local := local.(type) {
	case *locals.Block:
		for _, s := range local.Stmts {
			c.genLocal(s)
		}
	case *locals.Return:
		c.ensureBlock()
		c.genReturn(local)
	case *locals.Let:
		c.ensureBlock()
		c.genLocalLet(local)
	case *locals.If:
		c.ensureBlock()
		c.genIf(local)
	case *locals.While:
		c.ensureBlock()
		c.genWhile(local)
	case *locals.For:
		c.ensureBlock()
		c.genFor(local)
	case locals.Expr:
		c.ensureBlock()
		c.genExpr(local)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的语句 %T", local))
	}
}

// genIf if/else-if/else 语句（E3）：条件只求值一次，各分支跳转共享合流块。
//
// 若所有分支都以终结指令（return 等）结束，合流块将没有前驱；此时仍在合流块内补一条
// unreachable 作为终结指令（LLVM 要求每个基本块都必须有终结指令），并保持
// terminated=true，让后续语句经 ensureBlock 进入独立的死块。选择该方案而非"延迟创建
// 合流块"是因为：else-if 递归链上每层都要共享同一个合流块，延迟创建需要额外传递
// "是否已创建"状态；而空 unreachable 块无副作用，后续优化可直接删除。
func (c *CodeGenerator) genIf(l *locals.If) {
	endBlock := c.currentFunc.NewBlock("if.end")
	fellThrough := c.genIfChain(l, endBlock)
	c.moveTo(endBlock)
	if !fellThrough {
		c.builder.Unreachable()
		c.terminated = true
	}
}

// genIfChain 发射 if/else-if 链（else-if 递归），返回是否有分支末尾跳转到 endBlock
// （即合流块有前驱）。调用前插入点必须位于条件求值处。
func (c *CodeGenerator) genIfChain(l *locals.If, endBlock ir.Block) bool {
	cond := asInt(c.genExpr(l.Condition))
	thenBlock := c.currentFunc.NewBlock("if.then")
	elseBlock := c.currentFunc.NewBlock("if.else")
	c.builder.CondBr(cond, thenBlock, elseBlock)
	c.terminated = true

	c.moveTo(thenBlock)
	c.genLocal(l.Body)
	thenFell := !c.terminated
	if thenFell {
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(elseBlock)
	var elseFell bool
	if next, ok := l.Else.Value(); ok {
		if elseif, ok := next.TryLeft(); ok {
			// else-if：条件在 else 块内求值，合流块与上层共享
			elseFell = c.genIfChain(elseif, endBlock)
		} else {
			c.genLocal(next.Right())
			elseFell = !c.terminated
			if elseFell {
				c.builder.Br(endBlock)
				c.terminated = true
			}
		}
	} else {
		// 无 else：空 else 块直接跳合流块
		c.builder.Br(endBlock)
		c.terminated = true
		elseFell = true
	}
	return thenFell || elseFell
}

// genWhile while 循环（E4，语言中以 `for <cond> {}` 拼写）：cond→body→cond 块环；
// 循环体若以终结指令结束（如 return）则不再回跳。循环出口块始终有前驱（条件分支）。
func (c *CodeGenerator) genWhile(l *locals.While) {
	condBlock := c.currentFunc.NewBlock("while.cond")
	bodyBlock := c.currentFunc.NewBlock("while.body")
	endBlock := c.currentFunc.NewBlock("while.end")

	c.builder.Br(condBlock)
	c.terminated = true

	c.moveTo(condBlock)
	cond := asInt(c.genExpr(l.Condition))
	c.builder.CondBr(cond, bodyBlock, endBlock)
	c.terminated = true

	c.moveTo(bodyBlock)
	c.genLocal(l.Body)
	if !c.terminated {
		c.builder.Br(condBlock)
		c.terminated = true
	}

	c.moveTo(endBlock)
}

// genFor for x in range 遍历数组（E5）：i64 索引从 0 递增到数组长度，条件 `idx < size`。
// Range 只求值一次（不可寻址的字面量由 genAddrOrMaterialize 在入口块物化）；
// 每轮循环体开头把 range[idx] 存入 Var 的入口块 alloca。
func (c *CodeGenerator) genFor(l *locals.For) {
	at, ok := types.GetUnderlying(l.Range.GetType()).(types.ArrayType)
	if !ok {
		panic(fmt.Errorf("llgen: for-in 的遍历对象必须是数组（实际 %s）", l.Range.GetType()))
	}
	arrayT := c.genType(l.Range.GetType())
	// 先取 Range 地址（只求值一次，必要时物化），循环体按索引读取
	rangePtr := c.genAddrOrMaterialize(l.Range)
	if at.GetSize().Sign() <= 0 {
		// 零长数组（B10）：循环体一次也不执行
		return
	}

	i64 := c.ctx.LLVM().Int(64)
	idxPtr := c.allocaEntry(i64, "")
	c.builder.Store(i64.Const(0), idxPtr)
	varPtr := c.allocaEntry(c.genType(l.Var.GetType()), l.Var.GetName())
	c.ctx.idents[l.Var] = &Ident{Name: l.Var.GetName(), Local: varPtr}

	condBlock := c.currentFunc.NewBlock("for.cond")
	bodyBlock := c.currentFunc.NewBlock("for.body")
	endBlock := c.currentFunc.NewBlock("for.end")

	c.builder.Br(condBlock)
	c.terminated = true

	c.moveTo(condBlock)
	idx := c.builder.Load[llvm.IntT](idxPtr, i64, "")
	size := c.ctx.LLVM().ConstIntOfString(i64, at.GetSize().String(), 10)
	c.builder.CondBr(c.builder.ICmp(llvm.IntSLT, idx, size, ""), bodyBlock, endBlock)
	c.terminated = true

	c.moveTo(bodyBlock)
	elemT := c.genType(l.Var.GetType())
	if c.isZeroSizeLLVM(elemT) {
		// 零尺寸元素（B10）：元素无存储，直接给空聚合零值
		c.builder.Store(c.genZeroValue(elemT), varPtr)
	} else {
		elemPtr := c.builder.GEP(arrayT, rangePtr, c.gepPath(idx), "")
		c.builder.Store(c.builder.Load[llvm.DynT](elemPtr, elemT.DynType(), ""), varPtr)
	}
	c.genLocal(l.Body)
	if !c.terminated {
		c.builder.Store(c.builder.Add(idx, i64.Const(1), ""), idxPtr)
		c.builder.Br(condBlock)
		c.terminated = true
	}

	c.moveTo(endBlock)
}

// genReturn return 语句（E2）：先求值表达式（保留副作用），再按返回类型是否为 unit
// 选择 RetVoid/Ret。unit 函数里 `return g()`（g 为 unit 函数）或 `return (c ? g() : h())`
// 的值是 void call / br 伪值，必须走 RetVoid，否则生成非法 IR（`ret void <badref>`）。
func (c *CodeGenerator) genReturn(l *locals.Return) {
	if v, ok := l.Value.Value(); ok {
		val := c.genExpr(v)
		if _, ok := types.GetUnderlying(v.GetType()).(types.UnitType); ok {
			c.builder.RetVoid()
		} else {
			c.builder.Ret(val)
		}
	} else {
		// 非 unit 函数里的空 return 会生成非法 IR（Module.Verify 报 "Function return type
		// does not match operand type of return inst!"）。正常前端应拒绝，此处防御性报错，
		// 用后端错误替换难懂的验证器英文信息。
		if !c.currentFunc.Signature().Return().Equal(c.ctx.LLVM().Void()) {
			panic(fmt.Errorf("llgen: 非 unit 函数 %s 不能使用空 return（前端漏校验）", c.currentFunc.Name()))
		}
		c.builder.RetVoid()
	}
	c.terminated = true
}

func (c *CodeGenerator) genLocalLet(l *locals.Let) {
	// unit 类型不能 alloca（LLVM 不允许 void 存储）；正常前端会拒绝，此处防御性报错
	if _, ok := types.GetUnderlying(l.GetType()).(types.UnitType); ok {
		panic(fmt.Errorf("llgen: 不能为 unit 类型的变量 %s 分配存储（类型 %s）", l.Name, l.GetType()))
	}
	// 与参数一致在入口块分配（§4.3）：循环体内的 let 不会每轮消耗新栈空间
	ptr := c.allocaEntry(c.genType(l.GetType()), l.Name)
	c.ctx.idents[l] = &Ident{Name: l.Name, Local: ptr}
	if v, ok := l.Value.Value(); ok {
		c.builder.Store(c.genExpr(v), ptr)
	}
}
