package llgen

import (
	"fmt"

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
