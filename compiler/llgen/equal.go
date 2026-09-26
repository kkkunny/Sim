package llgen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// genEquals 相等/不等比较（G1~G5）：not=false 生成 ==，not=true 生成 !=，结果为 i1。
// 标量直接比较；聚合按需合成相等性辅助函数（G7 缓存去重）；语义与旧 C 后端 genEqual 对等。
// 辅助函数始终计算 ==，!= 在调用点取反（零尺寸与浮点 NaN 两种边界下二者等价）。
func (c *CodeGenerator) genEquals(not bool, t hir.Type, left, right llvm.AnyValue) llvm.AnyValue {
	return c.genEqualsAs(not, t, c.genType(t), left, right)
}

// genEqualsAs 带 LLVM 类型参数的相等比较：自定义类型递归解包底层语义，
// 但 LLVM 类型沿用自定义类型自身（named struct 与底层字面量结构体布局一致）。
func (c *CodeGenerator) genEqualsAs(not bool, t hir.Type, llvmT llvm.AnyType, left, right llvm.AnyValue) llvm.AnyValue {
	if ct, ok := t.(types.CustomType); ok {
		return c.genEqualsAs(not, ct.GetUnderlying(), llvmT, left, right)
	}
	eqOp := locals.BinaryOpEnum.Eq
	if not {
		eqOp = locals.BinaryOpEnum.Neq
	}
	switch t := t.(type) {
	case types.UnitType:
		// 零尺寸值：== 恒真、!= 恒假
		return c.ctx.LLVM().ConstBool(!not)
	case types.BooleanType:
		return c.builder.ICmp(intCmpPred(eqOp, true), asInt(left), asInt(right), "")
	case types.SintType, types.UintType:
		return c.builder.ICmp(intCmpPred(eqOp, true), asInt(left), asInt(right), "")
	case types.FloatType:
		// Eq→OEQ、Neq→UNE（NaN 语义与 C 一致，见 floatCmpPred）
		return c.builder.FCmp(floatCmpPred(eqOp), asFloat(left), asFloat(right), "")
	case types.RefType:
		return c.builder.ICmp(intCmpPred(eqOp, true), left, right, "")
	case types.StringType:
		panic(fmt.Errorf("llgen: 暂不支持 str 类型的相等比较（旧 C 后端同样不支持）"))
	case types.FuncType:
		// G6/F5 函数值相等属 M4
		panic(fmt.Errorf("llgen: 暂不支持函数值相等比较（G6/F5，M4）"))
	case types.ArrayType, types.TupleType, types.StructType, types.UnionType:
		if c.isZeroSizeLLVM(llvmT) {
			// 零尺寸类型（B10）恒真短路：== 为 true、!= 为 false
			return c.ctx.LLVM().ConstBool(!not)
		}
		fn := c.buildEqFunc(t, llvmT)
		res := c.builder.Call[llvm.IntT](fn, []llvm.AnyValue{left, right}, "")
		if not {
			return c.builder.Xor(res, c.ctx.LLVM().ConstBool(true), "")
		}
		return res
	default:
		panic(fmt.Errorf("llgen: 暂不支持 %s 类型的相等比较", t))
	}
}

// buildEqFunc 生成（或复用）类型 t 的相等性辅助函数（G7）：按类型键 + 模块标识缓存，
// 同类型在同一模块只生成一次。辅助函数设 internal 链接，跨包引用时各模块各自生成。
func (c *CodeGenerator) buildEqFunc(t hir.Type, llvmT llvm.AnyType) ir.Function {
	// 键含模块标识（避免跨模块复用同一 Function 句柄）与 LLVM/HIR 类型文本：自定义类型的
	// LLVM 名带包限定（stableName），跨包同名类型不会串用；字面量类型为结构等价文本。
	key := fmt.Sprintf("m%d|%s|%s", c.moduleID, llvmT.String(), t.String())
	if fn, ok := c.ctx.eqFuncs[key]; ok {
		return fn
	}
	boolT := c.ctx.LLVM().Bool()
	fn := c.module.NewFunction("_sim_eq_"+shortHash(key), c.ctx.LLVM().Fn(boolT, []llvm.AnyType{llvmT, llvmT}, false))
	fn.SetLinkage(llvm.LinkageInternal)
	// 先登记再生成函数体：体内递归请求同一键（理论上仅引用断开的递归类型）可直接复用
	c.ctx.eqFuncs[key] = fn
	c.genEqFuncBody(t, llvmT, fn)
	return fn
}

// shortHash 缓存键的短哈希（辅助函数命名，稳定且冲突概率可忽略）
func shortHash(key string) string {
	digest := sha256.Sum256([]byte(key))
	return hex.EncodeToString(digest[:8])
}

// genEqFuncBody 生成相等性辅助函数体（始终计算 ==）：
// 元组/结构体逐字段、数组逐元素短路比较；union 先比 tag 再按 tag switch 分派到载荷成员。
func (c *CodeGenerator) genEqFuncBody(t hir.Type, llvmT llvm.AnyType, fn ir.Function) {
	// 保存当前发射状态，函数体生成完毕后恢复到调用点
	prevFunc, prevTerminated := c.currentFunc, c.terminated
	prevBlock, hadBlock := c.builder.CurrentBlock()
	defer func() {
		c.currentFunc = prevFunc
		c.terminated = prevTerminated
		if hadBlock {
			c.builder.MoveToEnd(prevBlock)
		}
	}()

	c.currentFunc = fn
	c.moveTo(fn.NewBlock("entry"))
	x, y := fn.Param(0).Dyn(), fn.Param(1).Dyn()
	switch t := t.(type) {
	case types.ArrayType:
		c.genEqArrayBody(t, llvmT, x, y)
	case types.TupleType:
		c.genEqFieldsBody(x, y, t.GetElems())
	case types.StructType:
		c.genEqFieldsBody(x, y, stlslices.Map(t.GetFields(), func(_ int, f *types.StructField) hir.Type {
			return f.Type
		}))
	case types.UnionType:
		c.genEqUnionBody(t, llvmT, x, y)
	default:
		panic(fmt.Errorf("llgen: 相等性辅助函数不支持类型 %T", t))
	}
}

// genEqFieldsBody 元组/结构体的逐字段比较（G3/G4）：
// 不等即返回 false，全部相等返回 true（与旧 genEqual 的短路结构一致）。
func (c *CodeGenerator) genEqFieldsBody(x, y llvm.AnyValue, fieldTypes []hir.Type) {
	failBlock := c.currentFunc.NewBlock("eq.fail")
	trueBlock := c.currentFunc.NewBlock("eq.true")
	for i, ft := range fieldTypes {
		lv := c.builder.ExtractValue[llvm.DynT](x, []uint32{uint32(i)}, "")
		rv := c.builder.ExtractValue[llvm.DynT](y, []uint32{uint32(i)}, "")
		eq := c.genEquals(false, ft, lv, rv)
		if i == len(fieldTypes)-1 {
			c.builder.CondBr(asInt(eq), trueBlock, failBlock)
		} else {
			nextBlock := c.currentFunc.NewBlock("eq.next")
			c.builder.CondBr(asInt(eq), nextBlock, failBlock)
			c.terminated = true
			c.moveTo(nextBlock)
		}
	}
	if len(fieldTypes) == 0 {
		c.builder.Br(trueBlock)
	}
	c.terminated = true

	c.moveTo(failBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(false))
	c.terminated = true
	c.moveTo(trueBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(true))
	c.terminated = true
}

// genEqArrayBody 数组逐元素比较（G2）：i64 索引循环，发现不等立即返回 false。
func (c *CodeGenerator) genEqArrayBody(t types.ArrayType, llvmT llvm.AnyType, x, y llvm.AnyValue) {
	i64 := c.ctx.LLVM().Int(64)
	xs := c.allocaEntry(llvmT, "")
	ys := c.allocaEntry(llvmT, "")
	c.builder.Store(x, xs)
	c.builder.Store(y, ys)
	idxPtr := c.allocaEntry(i64, "")
	c.builder.Store(i64.Const(0), idxPtr)

	condBlock := c.currentFunc.NewBlock("eq.cond")
	bodyBlock := c.currentFunc.NewBlock("eq.body")
	trueBlock := c.currentFunc.NewBlock("eq.true")
	failBlock := c.currentFunc.NewBlock("eq.fail")
	c.builder.Br(condBlock)
	c.terminated = true

	c.moveTo(condBlock)
	idx := c.builder.Load[llvm.IntT](idxPtr, i64, "")
	size := c.ctx.LLVM().ConstIntOfString(i64, t.GetSize().String(), 10)
	c.builder.CondBr(c.builder.ICmp(llvm.IntSLT, idx, size, ""), bodyBlock, trueBlock)
	c.terminated = true

	c.moveTo(bodyBlock)
	arrT := llvm.AsArrayType(llvmT)
	lp := c.builder.GEP(arrT, xs, c.gepPath(idx), "")
	rp := c.builder.GEP(arrT, ys, c.gepPath(idx), "")
	elemT := c.genType(t.GetElem())
	lv := c.builder.Load[llvm.DynT](lp, elemT.DynType(), "")
	rv := c.builder.Load[llvm.DynT](rp, elemT.DynType(), "")
	eq := c.genEquals(false, t.GetElem(), lv, rv)
	nextBlock := c.currentFunc.NewBlock("eq.next")
	c.builder.CondBr(asInt(eq), nextBlock, failBlock)
	c.terminated = true

	c.moveTo(nextBlock)
	c.builder.Store(c.builder.Add(idx, i64.Const(1), ""), idxPtr)
	c.builder.Br(condBlock)
	c.terminated = true

	c.moveTo(failBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(false))
	c.terminated = true
	c.moveTo(trueBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(true))
	c.terminated = true
}

// genEqUnionBody union 比较（G5）：先比 tag（不同即不等），再按 tag switch 分派到载荷对应
// 成员比较；default 分支（非法 tag）按不等处理。载荷成员经 GEP payload 首字段读取。
func (c *CodeGenerator) genEqUnionBody(t types.UnionType, llvmT llvm.AnyType, x, y llvm.AnyValue) {
	i8 := c.ctx.LLVM().Int(8)
	xs := c.allocaEntry(llvmT, "")
	ys := c.allocaEntry(llvmT, "")
	c.builder.Store(x, xs)
	c.builder.Store(y, ys)
	xt := c.builder.Load[llvm.IntT](c.builder.GEP(llvmT, xs, c.gepPath(c.ctx.LLVM().Int(32).Const(0)), ""), i8, "")
	yt := c.builder.Load[llvm.IntT](c.builder.GEP(llvmT, ys, c.gepPath(c.ctx.LLVM().Int(32).Const(0)), ""), i8, "")

	dispatchBlock := c.currentFunc.NewBlock("eq.switch")
	trueBlock := c.currentFunc.NewBlock("eq.true")
	failBlock := c.currentFunc.NewBlock("eq.fail")
	c.builder.CondBr(c.builder.ICmp(llvm.IntEQ, xt, yt, ""), dispatchBlock, failBlock)
	c.terminated = true

	c.moveTo(dispatchBlock)
	sw := c.builder.Switch(xt, failBlock)
	elems := c.unionMemberTypes(t)
	payloadT, _ := c.genUnionPayload(elems)
	for i, et := range elems {
		caseBlock := c.currentFunc.NewBlock(fmt.Sprintf("eq.case%d", i+1))
		sw.AddCase(i8.Const(uint64(i)), caseBlock)
		c.moveTo(caseBlock)
		if et == nil {
			// 零尺寸成员（B10）：无载荷可比较，tag 相同即相等
			c.builder.Br(trueBlock)
			c.terminated = true
			continue
		}
		lv := c.builder.Load[llvm.DynT](c.unionPayloadMemberAddr(llvmT, payloadT, xs), et.DynType(), "")
		rv := c.builder.Load[llvm.DynT](c.unionPayloadMemberAddr(llvmT, payloadT, ys), et.DynType(), "")
		eq := c.genEquals(false, t.GetElems()[i], lv, rv)
		c.builder.CondBr(asInt(eq), trueBlock, failBlock)
		c.terminated = true
	}

	c.moveTo(failBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(false))
	c.terminated = true
	c.moveTo(trueBlock)
	c.builder.Ret(c.ctx.LLVM().ConstBool(true))
	c.terminated = true
}

// unionPayloadMemberAddr union 值指针 → payload 首字段（选定载荷成员）的地址
func (c *CodeGenerator) unionPayloadMemberAddr(uT, payloadT llvm.AnyType, slot llvm.Value[llvm.PtrT]) llvm.Value[llvm.PtrT] {
	payload := c.builder.GEP(uT, slot, c.gepPath(c.ctx.LLVM().Int(32).Const(1)), "")
	return c.builder.GEP(payloadT, payload, c.gepPath(c.ctx.LLVM().Int(32).Const(0)), "")
}
