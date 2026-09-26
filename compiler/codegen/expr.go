package codegen

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
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
	case *locals.DeRef:
		// 解引用：目标表达式的值即指针，装载其指向的值
		return c.genLoadFromAddr(expr)
	case locals.Unary:
		return c.genUnary(expr)
	case *locals.Binary:
		return c.genBinary(expr)
	case locals.Covert:
		return c.genCovert(expr)
	case *locals.Ternary:
		return c.genTernary(expr)
	case *locals.Tuple:
		return c.genTuple(expr)
	case *locals.Array:
		return c.genArray(expr)
	case *locals.Struct:
		return c.genStruct(expr)
	case *locals.TupleIndex, *locals.ArrayIndex, *locals.GetField:
		// 索引/字段访问：左值通道取地址后装载
		return c.genLoadFromAddr(expr)
	case *locals.Func:
		return c.genClosureValue(expr)
	case *locals.GetBind:
		return c.genGetBind(expr)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的表达式 %s（%T）", expr, expr))
	}
}

// genLoadFromAddr 以左值通道（genAddr）求值后再装载，供解引用/索引/字段访问等表达式复用
func (c *CodeGenerator) genLoadFromAddr(expr locals.Expr) llvm.AnyValue {
	return c.builder.Load[llvm.DynT](c.genAddr(expr), c.genType(expr.GetType()).DynType(), "")
}

// genAddr 左值通道：生成表达式对应的存储地址（D1/D11/D12）
func (c *CodeGenerator) genAddr(expr locals.Expr) llvm.Value[llvm.PtrT] {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		// 捕获变量优先（F7）：地址为当前包装函数的 ctx 字段（GEP）
		if ptr, ok := c.captureVars[expr.Define]; ok {
			return ptr
		}
		ident, ok := c.ctx.idents[expr.Define]
		if !ok {
			panic(fmt.Errorf("llgen: 未找到符号 %s", expr.Define.GetName()))
		}
		if !ident.Local.IsNil() {
			return ident.Local
		}
		if ident.FuncSymbol {
			panic(fmt.Errorf("llgen: 暂不支持函数符号取址 %s（函数值不是左值）", ident.Name))
		}
		// 全局变量：符号本身即存储地址（跨包时按需建外部声明）
		return c.getGlobalVar(ident.Name, c.genType(expr.GetType())).Value
	case *locals.DeRef:
		// 解引用表达式的值即指针
		return c.genExpr(expr.Target).Dyn().MustAs[llvm.PtrT]()
	case *locals.ArrayIndex:
		base := c.genAddrOrMaterialize(expr.From)
		at, ok := c.genType(expr.From.GetType()).(llvm.ArrayType)
		if !ok {
			// 零尺寸数组（B10）映射为 {}，元素没有实际存储；load/store 皆为空操作，
			// 返回零尺寸临时地址以维持左值语义
			return c.allocaEntry(c.genType(expr.GetType()), "")
		}
		return c.builder.GEP(at, base, c.gepPath(asInt(c.genExpr(expr.Index))), "")
	case *locals.TupleIndex:
		base := c.genAddrOrMaterialize(expr.From)
		st, ok := c.genType(expr.From.GetType()).(llvm.StructType)
		if !ok {
			panic(fmt.Errorf("llgen: 元组索引 %s 的基类型 %s 不是结构体", expr, expr.From.GetType()))
		}
		return c.builder.GEP(st, base, c.gepPath(c.ctx.LLVM().Int(32).Const(uint64(expr.Index.Int64()))), "")
	case *locals.GetField:
		return c.genFieldAddr(expr)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的左值表达式 %s（%T）", expr, expr))
	}
}

// genFieldAddr 字段左值地址（D12）：GEP [0, fieldIdx]；From 为引用时按自动解引用处理
func (c *CodeGenerator) genFieldAddr(expr *locals.GetField) llvm.Value[llvm.PtrT] {
	from := expr.From
	fromT := from.GetType()
	var base llvm.Value[llvm.PtrT]
	if rt, ok := fromT.(types.RefType); ok {
		// 自动解引用：From 的值即指向结构体的地址（analyze 通常已插入 DeRef，此处防御）
		fromT = rt.PtrTo()
		base = c.genExpr(from).Dyn().MustAs[llvm.PtrT]()
	} else {
		base = c.genAddrOrMaterialize(from)
	}
	structT, ok := types.GetUnderlying(fromT).(types.StructType)
	if !ok {
		panic(fmt.Errorf("llgen: 字段访问 %s 的基类型 %s 不是结构体", expr, fromT))
	}
	fieldIdx := -1
	for i, f := range structT.GetFields() {
		if f.Name == expr.Name {
			fieldIdx = i
			break
		}
	}
	if fieldIdx < 0 {
		panic(fmt.Errorf("llgen: 结构体 %s 中不存在字段 %s", fromT, expr.Name))
	}
	st, ok := c.genType(fromT).(llvm.StructType)
	if !ok {
		panic(fmt.Errorf("llgen: 结构体 %s 的 LLVM 表示不是结构体", fromT))
	}
	return c.builder.GEP(st, base, c.gepPath(c.ctx.LLVM().Int(32).Const(uint64(fieldIdx))), "")
}

// gepPath GEP 下标列表：[0, rest]（先索引聚合本身，再取其元素）
func (c *CodeGenerator) gepPath(rest llvm.ValueRef[llvm.IntT]) []llvm.ValueRef[llvm.IntT] {
	return []llvm.ValueRef[llvm.IntT]{c.ctx.LLVM().Int(32).Const(0), rest}
}

// genAddrOrMaterialize 获取表达式地址（§4.3）：左值直接取地址；
// 不可寻址时在入口块 alloca 后存入其值（后续 by-value 语义与旧后端一致）。
func (c *CodeGenerator) genAddrOrMaterialize(expr locals.Expr) llvm.Value[llvm.PtrT] {
	if c.addressable(expr) {
		return c.genAddr(expr)
	}
	ptr := c.allocaEntry(c.genType(expr.GetType()), "")
	c.builder.Store(c.genExpr(expr), ptr)
	return ptr
}

// addressable 表达式是否可直接取地址（与 genAddr 支持的分支保持一致）
func (c *CodeGenerator) addressable(expr locals.Expr) bool {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		if _, ok := c.captureVars[expr.Define]; ok {
			return true
		}
		ident, ok := c.ctx.idents[expr.Define]
		if !ok {
			return false
		}
		return !ident.Local.IsNil() || !ident.FuncSymbol
	case *locals.DeRef, *locals.ArrayIndex, *locals.TupleIndex, *locals.GetField:
		return true
	default:
		return false
	}
}

// genUnary 一元运算（D5）
func (c *CodeGenerator) genUnary(expr locals.Unary) llvm.AnyValue {
	switch expr := expr.(type) {
	case *locals.BitsReverse:
		// 按位取反；bool 的 ! 是 BooleanReverse，不走这里
		return c.builder.Not(asInt(c.genExpr(expr.GetOpTarget())), "")
	case *locals.BooleanReverse:
		return c.builder.Xor(asInt(c.genExpr(expr.GetOpTarget())), c.ctx.LLVM().ConstBool(true), "")
	case *locals.GetRef:
		return c.genAddr(expr.GetOpTarget())
	default:
		panic(fmt.Errorf("llgen: 暂不支持的一元表达式 %s（%T）", expr, expr))
	}
}

// asInt 断言值为整数（i1/i8/...），类型不符即 panic
func asInt(v llvm.AnyValue) llvm.Value[llvm.IntT] {
	return v.Dyn().MustAs[llvm.IntT]()
}

// asFloat 断言值为浮点，类型不符即 panic
func asFloat(v llvm.AnyValue) llvm.Value[llvm.FloatT] {
	return v.Dyn().MustAs[llvm.FloatT]()
}

// assignOp2BinaryOp 复合赋值的等价二元运算
var assignOp2BinaryOp = map[locals.BinaryOp]locals.BinaryOp{
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

// genBinary 二元运算：赋值/复合赋值（D6）、算术/位运算/移位（D3）
func (c *CodeGenerator) genBinary(expr *locals.Binary) llvm.AnyValue {
	if op, ok := assignOp2BinaryOp[expr.Op]; ok {
		return c.genAssignOp(expr, op)
	}
	if expr.Op == locals.BinaryOpEnum.Assign {
		return c.genAssign(expr)
	}

	switch expr.Op {
	case locals.BinaryOpEnum.Add, locals.BinaryOpEnum.Sub, locals.BinaryOpEnum.Mul, locals.BinaryOpEnum.Quo,
		locals.BinaryOpEnum.Rem, locals.BinaryOpEnum.And, locals.BinaryOpEnum.Or, locals.BinaryOpEnum.Xor,
		locals.BinaryOpEnum.Shl, locals.BinaryOpEnum.Shr:
		return c.genBinaryOp(expr.Op, expr.Left.GetType(), c.genExpr(expr.Left), c.genExpr(expr.Right))
	case locals.BinaryOpEnum.Eq, locals.BinaryOpEnum.Neq, locals.BinaryOpEnum.Lt, locals.BinaryOpEnum.Lte,
		locals.BinaryOpEnum.Gt, locals.BinaryOpEnum.Gte:
		return c.genCompare(expr.Op, expr.Left.GetType(), c.genExpr(expr.Left), c.genExpr(expr.Right))
	case locals.BinaryOpEnum.LogicAnd, locals.BinaryOpEnum.LogicOr:
		return c.genLogic(expr)
	default:
		// D14 自增不在本任务
		panic(fmt.Errorf("llgen: 暂不支持的二元运算 %s（%T，D14）", expr.Op, expr.Op))
	}
}

// genAssign 赋值：左值地址只求值一次，存储右值并返回该值（D6）
func (c *CodeGenerator) genAssign(expr *locals.Binary) llvm.AnyValue {
	addr := c.genAddr(expr.Left)
	val := c.genExpr(expr.Right)
	c.builder.Store(val, addr)
	return val
}

// genAssignOp 复合赋值：等价于 a = a op b，但左值地址只求值一次（D6）
func (c *CodeGenerator) genAssignOp(expr *locals.Binary, op locals.BinaryOp) llvm.AnyValue {
	addr := c.genAddr(expr.Left)
	old := c.builder.Load[llvm.DynT](addr, c.genType(expr.Left.GetType()).DynType(), "")
	right := c.genExpr(expr.Right)
	val := c.genBinaryOp(op, expr.Left.GetType(), old, right)
	c.builder.Store(val, addr)
	return val
}

// genBinaryOp 按操作数类型生成算术/位运算/移位指令（D3）；t 取左操作数类型
func (c *CodeGenerator) genBinaryOp(op locals.BinaryOp, t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	ut := types.GetUnderlying(t) // 自定义类型（typedef）递归解包到底层
	switch ut.(type) {
	case types.FloatType:
		l, r := asFloat(left), asFloat(right)
		switch op {
		case locals.BinaryOpEnum.Add:
			return c.builder.FAdd(l, r, "")
		case locals.BinaryOpEnum.Sub:
			return c.builder.FSub(l, r, "")
		case locals.BinaryOpEnum.Mul:
			return c.builder.FMul(l, r, "")
		case locals.BinaryOpEnum.Quo:
			return c.builder.FDiv(l, r, "")
		case locals.BinaryOpEnum.Rem:
			// 浮点取模用 FRem 指令，不调 libm
			return c.builder.FRem(l, r, "")
		}
	case types.SintType:
		return c.genIntBinaryOp(op, true, left, right)
	case types.UintType:
		return c.genIntBinaryOp(op, false, left, right)
	}
	panic(fmt.Errorf("llgen: 暂不支持的二元运算 %s（类型 %s，D4/D7）", op, t))
}

// genIntBinaryOp 整数算术/位运算/移位（D3）；signed 决定除法/取余/右移的符号性
func (c *CodeGenerator) genIntBinaryOp(op locals.BinaryOp, signed bool, left, right llvm.AnyValue) llvm.AnyValue {
	l, r := asInt(left), asInt(right)
	switch op {
	case locals.BinaryOpEnum.Add:
		return c.builder.Add(l, r, "")
	case locals.BinaryOpEnum.Sub:
		return c.builder.Sub(l, r, "")
	case locals.BinaryOpEnum.Mul:
		return c.builder.Mul(l, r, "")
	case locals.BinaryOpEnum.Quo:
		if signed {
			return c.builder.SDiv(l, r, "")
		}
		return c.builder.UDiv(l, r, "")
	case locals.BinaryOpEnum.Rem:
		if signed {
			return c.builder.SRem(l, r, "")
		}
		return c.builder.URem(l, r, "")
	case locals.BinaryOpEnum.And:
		return c.builder.And(l, r, "")
	case locals.BinaryOpEnum.Or:
		return c.builder.Or(l, r, "")
	case locals.BinaryOpEnum.Xor:
		return c.builder.Xor(l, r, "")
	case locals.BinaryOpEnum.Shl:
		return c.builder.Shl(l, r, "")
	case locals.BinaryOpEnum.Shr:
		if signed {
			return c.builder.AShr(l, r, "")
		}
		return c.builder.LShr(l, r, "")
	default:
		panic(fmt.Errorf("llgen: 暂不支持的整数二元运算 %s（D4/D7）", op))
	}
}

// genCompare 比较运算（D4/G1~G5）：结果为 i1；t 取左操作数类型。
// Eq/Neq 走 genEquals（复合类型按需合成辅助函数）；大小比较仅支持标量。
func (c *CodeGenerator) genCompare(op locals.BinaryOp, t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	if op == locals.BinaryOpEnum.Eq || op == locals.BinaryOpEnum.Neq {
		return c.genEquals(op == locals.BinaryOpEnum.Neq, t, left, right)
	}
	switch types.GetUnderlying(t).(type) {
	case types.FloatType:
		return c.builder.FCmp(floatCmpPred(op), asFloat(left), asFloat(right), "")
	case types.SintType:
		return c.builder.ICmp(intCmpPred(op, true), asInt(left), asInt(right), "")
	case types.UintType:
		return c.builder.ICmp(intCmpPred(op, false), asInt(left), asInt(right), "")
	case types.BooleanType:
		panic(fmt.Errorf("llgen: bool 不支持大小比较 %s", op))
	case types.RefType:
		panic(fmt.Errorf("llgen: 引用不支持大小比较 %s", op))
	default:
		panic(fmt.Errorf("llgen: 暂不支持 %s 类型的大小比较 %s", t, op))
	}
}

// intCmpPred 整数比较谓词（D4）
func intCmpPred(op locals.BinaryOp, signed bool) llvm.IntPred {
	switch op {
	case locals.BinaryOpEnum.Eq:
		return llvm.IntEQ
	case locals.BinaryOpEnum.Neq:
		return llvm.IntNE
	case locals.BinaryOpEnum.Lt:
		if signed {
			return llvm.IntSLT
		}
		return llvm.IntULT
	case locals.BinaryOpEnum.Lte:
		if signed {
			return llvm.IntSLE
		}
		return llvm.IntULE
	case locals.BinaryOpEnum.Gt:
		if signed {
			return llvm.IntSGT
		}
		return llvm.IntUGT
	case locals.BinaryOpEnum.Gte:
		if signed {
			return llvm.IntSGE
		}
		return llvm.IntUGE
	default:
		panic(fmt.Errorf("llgen: 非比较运算 %s", op))
	}
}

// floatCmpPred 浮点比较谓词（D4）：Eq/Lt/Lte/Gt/Gte 用有序谓词（任一操作数为 NaN 时结果恒假）。
// Neq 必须用 UNE（无序或不等）：C 的 `!=` 在任一操作数为 NaN 时返回 true，而 ONE 在 NaN 时返回 false。
func floatCmpPred(op locals.BinaryOp) llvm.FloatPred {
	switch op {
	case locals.BinaryOpEnum.Eq:
		return llvm.FloatOEQ
	case locals.BinaryOpEnum.Neq:
		return llvm.FloatUNE
	case locals.BinaryOpEnum.Lt:
		return llvm.FloatOLT
	case locals.BinaryOpEnum.Lte:
		return llvm.FloatOLE
	case locals.BinaryOpEnum.Gt:
		return llvm.FloatOGT
	case locals.BinaryOpEnum.Gte:
		return llvm.FloatOGE
	default:
		panic(fmt.Errorf("llgen: 非比较运算 %s", op))
	}
}

// genLogic 短路 && / ||（D7）：右侧惰性求值且最多执行一次，结果 i1
func (c *CodeGenerator) genLogic(expr *locals.Binary) llvm.AnyValue {
	boolT := c.ctx.LLVM().Bool()
	slot := c.allocaEntry(boolT, "")
	left := c.genExpr(expr.Left)
	c.builder.Store(left, slot)

	rhsBlock := c.currentFunc.NewBlock("logic.rhs")
	endBlock := c.currentFunc.NewBlock("logic.end")
	if expr.Op == locals.BinaryOpEnum.LogicAnd {
		// 左假短路
		c.builder.CondBr(asInt(left), rhsBlock, endBlock)
	} else {
		// 左真短路
		c.builder.CondBr(asInt(left), endBlock, rhsBlock)
	}
	c.terminated = true

	c.moveTo(rhsBlock)
	right := c.genExpr(expr.Right)
	if !c.terminated {
		c.builder.Store(right, slot)
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(endBlock)
	return c.builder.Load[llvm.DynT](slot, boolT.DynType(), "")
}

// genTernary 三元 ?:（D8）：只执行命中的分支（分支可有副作用），禁止 Select 双求值
func (c *CodeGenerator) genTernary(expr *locals.Ternary) llvm.AnyValue {
	resultT := c.genType(expr.GetType())
	if _, ok := resultT.(llvm.VoidType); ok {
		return c.genTernaryVoid(expr)
	}

	slot := c.allocaEntry(resultT, "")
	cond := c.genExpr(expr.Condition)

	thenBlock := c.currentFunc.NewBlock("tern.then")
	elseBlock := c.currentFunc.NewBlock("tern.else")
	endBlock := c.currentFunc.NewBlock("tern.end")
	c.builder.CondBr(asInt(cond), thenBlock, elseBlock)
	c.terminated = true

	c.moveTo(thenBlock)
	trueV := c.genExpr(expr.TrueExpr)
	if !c.terminated {
		c.builder.Store(trueV, slot)
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(elseBlock)
	falseV := c.genExpr(expr.FalseExpr)
	if !c.terminated {
		c.builder.Store(falseV, slot)
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(endBlock)
	return c.builder.Load[llvm.DynT](slot, resultT.DynType(), "")
}

// genTernaryVoid unit 结果的三元：无合流值，只保留分支结构
func (c *CodeGenerator) genTernaryVoid(expr *locals.Ternary) llvm.AnyValue {
	cond := asInt(c.genExpr(expr.Condition))

	thenBlock := c.currentFunc.NewBlock("tern.then")
	elseBlock := c.currentFunc.NewBlock("tern.else")
	endBlock := c.currentFunc.NewBlock("tern.end")
	br := c.builder.CondBr(cond, thenBlock, elseBlock)
	c.terminated = true

	c.moveTo(thenBlock)
	c.genExpr(expr.TrueExpr)
	if !c.terminated {
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(elseBlock)
	c.genExpr(expr.FalseExpr)
	if !c.terminated {
		c.builder.Br(endBlock)
		c.terminated = true
	}

	c.moveTo(endBlock)
	return br.Dyn()
}

// allocaEntry 在函数入口块分配存储（局部变量、短路/三元的合流槽）：入口块支配所有可达块，
// 且不会随循环体每轮重复分配导致栈无限增长。
func (c *CodeGenerator) allocaEntry(t llvm.AnyType, name string) llvm.Value[llvm.PtrT] {
	cur, ok := c.builder.CurrentBlock()
	entry, ok2 := c.currentFunc.EntryBlock()
	if !ok || !ok2 {
		return c.builder.Alloca(t, name).Value
	}
	// 入口块已插入指令则插入到首指令前（保证支配所有可达分支），否则直接追加
	if first, ok := entry.FirstInst(); ok {
		c.builder.MoveBefore(first)
	} else {
		c.builder.MoveToEnd(entry)
	}
	ptr := c.builder.Alloca(t, name).Value
	// 仅恢复插入位置，不改变 terminated 状态
	c.builder.MoveToEnd(cur)
	return ptr
}

// genCovert 转换表达式（D9 数值转换 / D10 typedef / B9 union 注入）
func (c *CodeGenerator) genCovert(expr locals.Covert) llvm.AnyValue {
	switch expr := expr.(type) {
	case *locals.NumberCovert:
		return c.genNumberConvert(
			types.GetUnderlying(expr.GetFrom().GetType()),
			types.GetUnderlying(expr.GetType()),
			c.genExpr(expr.GetFrom()),
		)
	case *locals.TypedefCovert:
		from, to := expr.GetFrom().GetType(), expr.GetType()
		v := c.genExpr(expr.GetFrom())
		if c.genType(from).Equal(c.genType(to)) {
			// 标量别名/str/引用别名等底层 LLVM 类型一致：直接透传
			return v
		}
		fromU, toU := types.GetUnderlying(from), types.GetUnderlying(to)
		if isNumber(fromU) && isNumber(toU) {
			return c.genNumberConvert(fromU, toU, v)
		}
		// 聚合层（D10）：analyze 仅在 GetUnderlying(from).Equal(GetUnderlying(to)) 时构造，
		// 布局相同（如两个同布局的命名聚合类型）；经内存往返完成位重解释
		slot := c.allocaEntry(c.genType(from), "")
		c.builder.Store(v, slot)
		return c.builder.Load[llvm.DynT](slot, c.genType(to).DynType(), "")
	case *locals.Union:
		return c.genUnionInject(expr)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的转换表达式 %s（%T）", expr, expr))
	}
}

// genUnionInject union 注入（B9）：tag = 成员下标，载荷成员写入 payload 首字段（偏移 0），
// 其余字节先清零（C 的 designated initializer 语义），最后 InsertValue 组成 union 值。
func (c *CodeGenerator) genUnionInject(expr *locals.Union) llvm.AnyValue {
	val := c.genExpr(expr.GetFrom())
	uT := c.genType(expr.GetType())
	st, ok := uT.(llvm.StructType)
	if !ok {
		panic(fmt.Errorf("llgen: union 类型 %s 的 LLVM 表示不是结构体", expr.GetType()))
	}
	if c.isZeroSizeLLVM(st) {
		// 全零尺寸 union → {}（B10）：无 tag 与载荷
		return c.genZeroValue(st)
	}
	unionT, ok := types.GetUnderlying(expr.GetType()).(types.UnionType)
	if !ok {
		panic(fmt.Errorf("llgen: union 注入的目标类型 %s 不是 union", expr.GetType()))
	}
	elems := c.unionMemberTypes(unionT)
	if int(expr.Index) >= len(elems) {
		panic(fmt.Errorf("llgen: union %s 的成员下标 %d 越界", expr.GetType(), expr.Index))
	}
	payloadT, _ := c.genUnionPayload(elems)
	slot := c.allocaEntry(payloadT, "")
	c.builder.Store(c.genZeroValue(payloadT), slot)
	if elems[expr.Index] != nil {
		field := c.builder.GEP(payloadT, slot, c.gepPath(c.ctx.LLVM().Int(32).Const(0)), "")
		c.builder.Store(val, field)
	}
	payload := c.builder.Load[llvm.DynT](slot, payloadT.DynType(), "")
	tag := c.ctx.LLVM().Int(8).Const(uint64(expr.Index))
	base := st.Zero()
	withPayload := c.builder.InsertValue[llvm.StructT](base, payload, []uint32{1}, "")
	return c.builder.InsertValue[llvm.StructT](withPayload, tag, []uint32{0}, "")
}

// genNumberConvert 标量数值转换（D9）：from/to 均为已解包的基础类型
func (c *CodeGenerator) genNumberConvert(from, to hir.Type, v llvm.AnyValue) llvm.AnyValue {
	switch from := from.(type) {
	case types.SintType:
		return c.genIntConvert(v, from.GetBits(), true, to)
	case types.UintType:
		return c.genIntConvert(v, from.GetBits(), false, to)
	case types.FloatType:
		return c.genFloatConvert(v, from.GetBits(), to)
	default:
		panic(fmt.Errorf("llgen: 不支持的数值转换 %s -> %s", from, to))
	}
}

// genIntConvert 整数 → 整数/浮点（D9）
func (c *CodeGenerator) genIntConvert(v llvm.AnyValue, fromBits uint8, signed bool, to hir.Type) llvm.AnyValue {
	iv := asInt(v)
	if ft, ok := to.(types.FloatType); ok {
		dst := c.genType(ft).(llvm.FloatType)
		if signed {
			return c.builder.SIToFP(iv, dst, "")
		}
		return c.builder.UIToFP(iv, dst, "")
	}
	toNum, ok := to.(types.NumberType)
	if !ok {
		panic(fmt.Errorf("llgen: 不支持的整数目标类型 %s", to))
	}
	dst := c.genType(to).(llvm.IntType)
	switch toBits := toNum.GetBits(); {
	case toBits < fromBits:
		return c.builder.Trunc(iv, dst, "")
	case toBits > fromBits:
		if signed {
			return c.builder.SExt(iv, dst, "")
		}
		return c.builder.ZExt(iv, dst, "")
	default:
		// 同宽：仅符号性变化，位模式不变
		return iv
	}
}

// genFloatConvert 浮点 → 浮点/整数（D9）
func (c *CodeGenerator) genFloatConvert(v llvm.AnyValue, fromBits uint8, to hir.Type) llvm.AnyValue {
	fv := asFloat(v)
	switch to := to.(type) {
	case types.FloatType:
		dst := c.genType(to).(llvm.FloatType)
		switch toBits := to.GetBits(); {
		case toBits < fromBits:
			return c.builder.FPTrunc(fv, dst, "")
		case toBits > fromBits:
			return c.builder.FPExt(fv, dst, "")
		default:
			return fv
		}
	case types.SintType:
		return c.builder.FPToSI(fv, c.genType(to).(llvm.IntType), "")
	case types.UintType:
		return c.builder.FPToUI(fv, c.genType(to).(llvm.IntType), "")
	default:
		panic(fmt.Errorf("llgen: 不支持的浮点目标类型 %s", to))
	}
}

// isNumber 是否标量数值类型
func isNumber(t hir.Type) bool {
	switch t.(type) {
	case types.SintType, types.UintType, types.FloatType:
		return true
	default:
		return false
	}
}

func (c *CodeGenerator) genIdentExpr(expr *locals.IdentExpr) llvm.AnyValue {
	// 捕获变量优先（F7）：当前包装函数中从 ctx 字段装载
	if ptr, ok := c.captureVars[expr.Define]; ok {
		return c.builder.Load[llvm.DynT](ptr, c.genType(expr.GetType()).DynType(), "")
	}
	ident, ok := c.ctx.idents[expr.Define]
	if !ok {
		panic(fmt.Errorf("llgen: 未找到符号 %s", expr.Define.GetName()))
	}
	if !ident.Local.IsNil() {
		return c.builder.Load[llvm.DynT](ident.Local, c.genType(expr.GetType()).DynType(), "")
	}
	if ident.FuncSymbol {
		// 全局函数/外部函数作为值（F3）：{fnptr, null}
		ft, ok := asFuncType(expr.GetType())
		if !ok {
			panic(fmt.Errorf("llgen: 函数符号 %s 的类型 %s 不是函数类型", ident.Name, expr.GetType()))
		}
		return c.genFuncSymbolValue(ident, ft)
	}
	// 全局变量：跨包/后向引用时按需在本模块创建 external 声明
	t := c.genType(expr.GetType())
	g := c.getGlobalVar(ident.Name, t)
	return c.builder.Load[llvm.DynT](g.Value, t.DynType(), "")
}

// asFuncType 取函数类型并递归解包自定义类型别名
func asFuncType(t hir.Type) (types.FuncType, bool) {
	ft, ok := types.GetUnderlying(t).(types.FuncType)
	return ft, ok
}

// genTuple 元组字面量（C3）：运行时按元素顺序求值并构造
func (c *CodeGenerator) genTuple(expr *locals.Tuple) llvm.AnyValue {
	return c.genAggregateLiteral(c.genType(expr.GetType()), stlslices.Map(expr.Elems, func(_ int, e locals.Expr) llvm.AnyValue {
		return c.genExpr(e)
	}))
}

// genArray 数组字面量（C3）：运行时按元素顺序求值并构造
func (c *CodeGenerator) genArray(expr *locals.Array) llvm.AnyValue {
	return c.genAggregateLiteral(c.genType(expr.GetType()), stlslices.Map(expr.Elems, func(_ int, e locals.Expr) llvm.AnyValue {
		return c.genExpr(e)
	}))
}

// genStruct 结构体字面量（C3）：按类型字段顺序求值并构造（Fields 是 map，不能直接遍历）。
// 缺省字段按 C designated initializer 语义零初始化（旧后端同样依赖该语义）。
func (c *CodeGenerator) genStruct(expr *locals.Struct) llvm.AnyValue {
	fields := expr.Type.GetFields()
	elems := make([]llvm.AnyValue, len(fields))
	for i, f := range fields {
		if fv, ok := expr.Fields[f.Name]; ok {
			elems[i] = c.genExpr(fv)
		} else {
			elems[i] = c.genZeroValue(c.genType(f.Type))
		}
	}
	return c.genAggregateLiteral(c.genType(expr.Type), elems)
}

// genAggregateLiteral 运行时构造聚合值：入口块 alloca + 逐元素 GEP/store + 整体 load。
// 元组/数组/结构体字面量统一走该路径（元素副作用顺序与书写顺序一致）；
// 零尺寸字面量直接给出空聚合零值，不走内存往返。
func (c *CodeGenerator) genAggregateLiteral(t llvm.AnyType, elems []llvm.AnyValue) llvm.AnyValue {
	if len(elems) == 0 || c.isZeroSizeLLVM(t) {
		return c.genZeroValue(t)
	}
	slot := c.allocaEntry(t, "")
	for i, e := range elems {
		ptr := c.builder.GEP(t, slot, c.gepPath(c.ctx.LLVM().Int(32).Const(uint64(i))), "")
		c.builder.Store(e, ptr)
	}
	return c.builder.Load[llvm.DynT](slot, t.DynType(), "")
}

// genConstExpr 常量路径（C3 全局初始化用）：Integer/Float/Boolean/String 及
// Tuple/Array/Struct 的常量嵌套；不在常量集返回 false（调用方按语义报错）。
func (c *CodeGenerator) genConstExpr(expr locals.Expr) (llvm.AnyValue, bool) {
	switch expr := expr.(type) {
	case *locals.Integer:
		it, ok := c.genType(expr.GetType()).(llvm.IntType)
		if !ok {
			return nil, false
		}
		return c.ctx.LLVM().ConstIntOfString(it, expr.Value.String(), 10), true
	case *locals.Float:
		ft, ok := c.genType(expr.GetType()).(llvm.FloatType)
		if !ok {
			return nil, false
		}
		v, _ := expr.Value.Float64()
		return c.ctx.LLVM().ConstFloat(ft, v), true
	case *locals.Boolean:
		return c.ctx.LLVM().ConstBool(expr.Value), true
	case *locals.String:
		// 字符串常量内部走 genString 的全局常量构造（ConstGEP + ConstNamedStruct）
		return c.genString(expr), true
	case *locals.IdentExpr:
		// 全局函数/外部函数符号作为常量函数值（{fnptr, null}）：
		// 函数指针是常量，可作为全局变量初始化器（B8/F3）
		ident, ok := c.ctx.idents[expr.Define]
		if !ok || !ident.FuncSymbol {
			return nil, false
		}
		ft, ok := asFuncType(expr.GetType())
		if !ok {
			return nil, false
		}
		fn := c.getFunction(ident.Name, c.genNativeFuncType(ft))
		return c.packNullFuncValue(fn.AsValue()), true
	case *locals.Tuple:
		return c.genConstAggregate(c.genType(expr.GetType()), expr.Elems)
	case *locals.Array:
		return c.genConstAggregate(c.genType(expr.GetType()), expr.Elems)
	case *locals.Struct:
		fields := expr.Type.GetFields()
		elems := make([]llvm.AnyValue, len(fields))
		for i, f := range fields {
			fv, ok := expr.Fields[f.Name]
			if !ok {
				// 缺省字段零初始化（与运行时构造一致）
				elems[i] = c.genZeroValue(c.genType(f.Type))
				continue
			}
			v, ok := c.genConstExpr(fv)
			if !ok {
				return nil, false
			}
			elems[i] = v
		}
		return c.genConstValues(c.genType(expr.Type), elems)
	default:
		return nil, false
	}
}

// genConstAggregate 聚合常量：逐元素递归求常量后按 LLVM 结构/数组类型构造
func (c *CodeGenerator) genConstAggregate(t llvm.AnyType, exprs []locals.Expr) (llvm.AnyValue, bool) {
	if len(exprs) == 0 || c.isZeroSizeLLVM(t) {
		return c.genZeroValue(t), true
	}
	elems := make([]llvm.AnyValue, len(exprs))
	for i, e := range exprs {
		v, ok := c.genConstExpr(e)
		if !ok {
			return nil, false
		}
		elems[i] = v
	}
	return c.genConstValues(t, elems)
}

// genConstValues 已知各元素常量时按 LLVM 聚合类型构造常量
func (c *CodeGenerator) genConstValues(t llvm.AnyType, elems []llvm.AnyValue) (llvm.AnyValue, bool) {
	if len(elems) == 0 || c.isZeroSizeLLVM(t) {
		return c.genZeroValue(t), true
	}
	switch t := t.(type) {
	case llvm.StructType:
		if t.Name() != "" {
			return c.ctx.LLVM().ConstNamedStruct(t, elems...), true
		}
		return c.ctx.LLVM().ConstStruct(false, elems...), true
	case llvm.ArrayType:
		return c.ctx.LLVM().ConstArray(t.Elem(), elems...), true
	default:
		return nil, false
	}
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
	// 字符串常量仅本模块引用；每个模块都从 1 开始计数，若用默认 external 链接，
	// 跨模块链接会报 multiple definition，故设为 private（不进符号表）
	g.SetLinkage(llvm.LinkagePrivate)
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

// genCall 调用（D13/F4）：静态已知形态走直接调用快速路径，其余函数值走运行时
// ctx 判空双分支间接调用。
func (c *CodeGenerator) genCall(expr *locals.Call) llvm.AnyValue {
	ft, ok := asFuncType(expr.Func.GetType())
	if !ok {
		panic(fmt.Errorf("llgen: 被调用表达式 %s 的类型 %s 不是函数类型", expr.Func, expr.Func.GetType()))
	}
	// 快速路径 1：被调方是函数符号（全局函数/外部函数），直接 Call
	if identExpr, ok := expr.Func.(*locals.IdentExpr); ok {
		if ident, ok := c.ctx.idents[identExpr.Define]; ok && ident.FuncSymbol {
			fn := c.getFunction(ident.Name, c.genNativeFuncType(ft))
			return c.builder.Call[llvm.DynT](fn, c.genCallArgs(expr), "")
		} else if !ok {
			panic(fmt.Errorf("llgen: 符号 %s 尚未登记（依赖模块需先生成）", identExpr.Define.GetName()))
		}
	}
	// 快速路径 2：被调方是函数字面量，fn 静态已知
	if fnLit, ok := expr.Func.(*locals.Func); ok {
		info := c.buildClosure(fnLit)
		args := c.genCallArgs(expr)
		if len(info.captures) == 0 {
			return c.builder.Call[llvm.DynT](info.fn, args, "")
		}
		// 有捕获：定义点构造 ctx 后直接以 ctx 调用包装函数（旧后端同为直接调用）
		ctxPtr := c.buildClosureCtx(info)
		return c.builder.CallIndirect[llvm.DynT](info.fn.AsValue().MustAs[llvm.PtrT](), c.genCtxFuncType(ft),
			append([]llvm.AnyValue{ctxPtr}, args...), "")
	}
	// 一般函数值（变量/参数/字段/返回的闭包/GetBind）：运行时分支
	return c.genCallValue(expr, ft)
}

// genCallArgs 求值实参列表（副作用顺序与书写顺序一致，且只求值一次）
func (c *CodeGenerator) genCallArgs(expr *locals.Call) []llvm.AnyValue {
	return stlslices.Map(expr.Args, func(_ int, e locals.Expr) llvm.AnyValue {
		return c.genExpr(e)
	})
}

// genCallValue 一般函数值调用（F4/§4.4）：被调方求值一次 → 提取 fn/ctx →
// `ctx == null ? fn(args...) : fn(ctx, args...)` 两个 CallIndirect + 合流。
// 实参在分支之前求值一次，保证副作用次数与直接调用一致；unit 返回无合流值。
func (c *CodeGenerator) genCallValue(expr *locals.Call, ft types.FuncType) llvm.AnyValue {
	fat := c.genExpr(expr.Func)
	fn := c.builder.ExtractValue[llvm.PtrT](fat, []uint32{0}, "")
	ctxPtr := c.builder.ExtractValue[llvm.PtrT](fat, []uint32{1}, "")
	args := c.genCallArgs(expr)
	isNull := c.builder.ICmp(llvm.IntEQ, ctxPtr, c.ctx.LLVM().Ptr(0).Null(), "")

	directBlock := c.currentFunc.NewBlock("call.direct")
	closureBlock := c.currentFunc.NewBlock("call.closure")
	endBlock := c.currentFunc.NewBlock("call.end")
	br := c.builder.CondBr(isNull, directBlock, closureBlock)
	c.terminated = true

	retT := c.genType(ft.GetReturn())
	retVoid := isVoidType(retT)
	var slot llvm.Value[llvm.PtrT]
	if !retVoid {
		slot = c.allocaEntry(retT, "")
	}

	// ctx == null：普通函数值，调用无 ctx 签名
	c.moveTo(directBlock)
	direct := c.builder.CallIndirect[llvm.DynT](fn, c.genNativeFuncType(ft), args, "")
	if !retVoid {
		c.builder.Store(direct, slot)
	}
	c.builder.Br(endBlock)
	c.terminated = true

	// ctx != null：闭包，首参传 ctx 调用包装函数
	c.moveTo(closureBlock)
	closure := c.builder.CallIndirect[llvm.DynT](fn, c.genCtxFuncType(ft), append([]llvm.AnyValue{ctxPtr}, args...), "")
	if !retVoid {
		c.builder.Store(closure, slot)
	}
	c.builder.Br(endBlock)
	c.terminated = true

	c.moveTo(endBlock)
	if retVoid {
		// unit 无合流值；返回伪值仅供语句位置忽略（与 genTernaryVoid 同做法）
		return br.Dyn()
	}
	return c.builder.Load[llvm.DynT](slot, retT.DynType(), "")
}

// isVoidType 是否 void（unit）
func isVoidType(t llvm.AnyType) bool {
	_, ok := t.(llvm.VoidType)
	return ok
}

func (c *CodeGenerator) getFunction(name string, sig llvm.FnType) ir.Function {
	if fn, ok := c.module.GetFunction(name); ok {
		return fn
	}
	return c.module.NewFunction(name, sig)
}
