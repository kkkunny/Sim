package llgen

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
		return c.builder.Load[llvm.DynT](c.genAddr(expr), c.genType(expr.GetType()).DynType(), "")
	case locals.Unary:
		return c.genUnary(expr)
	case *locals.Binary:
		return c.genBinary(expr)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的表达式 %s（%T）", expr, expr))
	}
}

// genAddr 左值通道：生成表达式对应的存储地址（D1）
func (c *CodeGenerator) genAddr(expr locals.Expr) llvm.Value[llvm.PtrT] {
	switch expr := expr.(type) {
	case *locals.IdentExpr:
		ident, ok := c.ctx.idents[expr.Define]
		if !ok {
			panic(fmt.Errorf("llgen: 未找到符号 %s", expr.Define.GetName()))
		}
		if ident.Local.IsNil() {
			panic(fmt.Errorf("llgen: 符号 %s 不是可寻址的局部变量（C4/F3）", ident.Name))
		}
		return ident.Local
	case *locals.DeRef:
		// 解引用表达式的值即指针
		return c.genExpr(expr.Target).Dyn().MustAs[llvm.PtrT]()
	case *locals.ArrayIndex:
		panic(fmt.Errorf("llgen: 暂不支持数组索引左值（%T，待 B4/B5/B6 实现）", expr))
	case *locals.TupleIndex:
		panic(fmt.Errorf("llgen: 暂不支持元组索引左值（%T，待 B4/B5/B6 实现）", expr))
	case *locals.GetField:
		panic(fmt.Errorf("llgen: 暂不支持字段左值（%T，待 B4/B5/B6 实现）", expr))
	default:
		panic(fmt.Errorf("llgen: 暂不支持的左值表达式 %s（%T）", expr, expr))
	}
}

// genAddrOrMaterialize 获取表达式地址：左值直接取地址；非左值后续任务物化临时存储
func (c *CodeGenerator) genAddrOrMaterialize(expr locals.Expr) llvm.Value[llvm.PtrT] {
	return c.genAddr(expr)
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
	default:
		// D4 比较、D7 短路、D14 自增不在本任务
		panic(fmt.Errorf("llgen: 暂不支持的二元运算 %s（%T，D4/D7）", expr.Op, expr.Op))
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
