package codegen

import (
	"unsafe"

	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
	stlutil "github.com/kkkunny/stl/util"
)

// 表达式
func (self *CodeGenerator) codegenExpr(mean hir.Expr, getValue bool) llvm.Value {
	switch expr := mean.(type) {
	case *hir.Integer, *hir.Float, *hir.Boolean, *hir.String, *hir.EmptyFunc, *hir.EmptyPtr, *hir.EmptyStruct, *hir.EmptyArray, *hir.EmptyTuple:
		return self.codegenConstantExpr(mean)
	case hir.Ident:
		switch ident := expr.(type) {
		case *hir.Param:
			v := self.vars[ident]
			if getValue {
				v = self.builder.CreateLoad(v, "")
			}
			return v
		case *hir.Function:
			return self.vars[ident]
		case *hir.Method:
			return self.vars[ident]
		case *hir.Variable:
			v := self.vars[ident]
			if getValue {
				v = self.builder.CreateLoad(v, "")
			}
			return v
		case *hir.GlobalValue:
			v := self.vars[ident]
			if getValue {
				v = self.builder.CreateLoad(v, "")
			}
			return v
		default:
			panic("unreachable")
		}
	case hir.Binary:
		switch binary := expr.(type) {
		case *hir.Assign:
			left, right := self.codegenExpr(binary.Left, false), self.codegenExpr(binary.Right, true)
			self.builder.CreateStore(right, left)
			return llvm.Value{}
		case *hir.Equal:
			left, right := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			return self.builder.CreateIntCast(self.equal(left, right), t_bool, "")
		case *hir.NotEqual:
			left, right := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			left = self.equal(left, right)
			v := self.builder.CreateXor(left, llvm.ConstInt(left.Type(), 1, true), "")
			return self.builder.CreateIntCast(v, t_bool, "")
		case *hir.Lt:
			left, right := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			var v llvm.Value
			if binary.Left.Type().IsSint() {
				v = self.builder.CreateICmp(llvm.IntSLT, left, right, "")
			} else if binary.Left.Type().IsUint() {
				v = self.builder.CreateICmp(llvm.IntULT, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOLT, left, right, "")
			}
			return self.builder.CreateIntCast(v, t_bool, "")
		case *hir.Le:
			left, right := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			var v llvm.Value
			if binary.Left.Type().IsSint() {
				v = self.builder.CreateICmp(llvm.IntSLE, left, right, "")
			} else if binary.Left.Type().IsUint() {
				v = self.builder.CreateICmp(llvm.IntULE, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOLE, left, right, "")
			}
			return self.builder.CreateIntCast(v, t_bool, "")
		case *hir.Add:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateNSWAdd(l, r, "")
			} else if binary.Left.Type().IsUint() {
				return self.builder.CreateNUWAdd(l, r, "")
			} else {
				return self.builder.CreateFAdd(l, r, "")
			}
		case *hir.Sub:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateNSWSub(l, r, "")
			} else if binary.Left.Type().IsUint() {
				return self.builder.CreateNUWSub(l, r, "")
			} else {
				return self.builder.CreateFSub(l, r, "")
			}
		case *hir.Mul:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateNSWMul(l, r, "")
			} else if binary.Left.Type().IsUint() {
				return self.builder.CreateNUWMul(l, r, "")
			} else {
				return self.builder.CreateFMul(l, r, "")
			}
		case *hir.Div:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateSDiv(l, r, "")
			} else if binary.Left.Type().IsUint() {
				return self.builder.CreateUDiv(l, r, "")
			} else {
				return self.builder.CreateFDiv(l, r, "")
			}
		case *hir.Mod:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateSRem(l, r, "")
			} else if binary.Left.Type().IsUint() {
				return self.builder.CreateURem(l, r, "")
			} else {
				return self.builder.CreateFRem(l, r, "")
			}
		case *hir.And:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			return self.builder.CreateAnd(l, r, "")
		case *hir.Or:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			return self.builder.CreateOr(l, r, "")
		case *hir.Xor:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			return self.builder.CreateXor(l, r, "")
		case *hir.Shl:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			return self.builder.CreateShl(l, r, "")
		case *hir.Shr:
			l, r := self.codegenExpr(binary.Left, true), self.codegenExpr(binary.Right, true)
			if binary.Left.Type().IsSint() {
				return self.builder.CreateAShr(l, r, "")
			} else {
				return self.builder.CreateLShr(l, r, "")
			}
		case *hir.LogicAnd:
			nb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
			self.builder.CreateCondBr(
				self.builder.CreateIntCast(
					self.codegenExpr(binary.Left, true),
					self.ctx.Int1Type(),
					"",
				), nb, eb,
			)
			pb := self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(nb)
			nv := self.builder.CreateIntCast(self.codegenExpr(binary.Right, true), self.ctx.Int1Type(), "")
			self.builder.CreateBr(eb)
			nb = self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(eb)
			phi := self.builder.CreatePHI(nv.Type(), "")
			phi.AddIncoming([]llvm.Value{llvm.ConstInt(self.ctx.Int1Type(), 0, true), nv}, []llvm.BasicBlock{pb, nb})
			return self.builder.CreateIntCast(phi, t_bool, "")
		case *hir.LogicOr:
			nb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
			self.builder.CreateCondBr(
				self.builder.CreateIntCast(
					self.codegenExpr(binary.Left, true),
					self.ctx.Int1Type(),
					"",
				), eb, nb,
			)
			pb := self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(nb)
			nv := self.builder.CreateIntCast(self.codegenExpr(binary.Right, true), self.ctx.Int1Type(), "")
			self.builder.CreateBr(eb)
			nb = self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(eb)
			phi := self.builder.CreatePHI(nv.Type(), "")
			phi.AddIncoming([]llvm.Value{llvm.ConstInt(self.ctx.Int1Type(), 1, true), nv}, []llvm.BasicBlock{pb, nb})
			return self.builder.CreateIntCast(phi, t_bool, "")
		default:
			panic("unreachable")
		}
	case hir.Call:
		switch call := expr.(type) {
		case *hir.FuncCall:
			f := self.codegenExpr(call.Func, true)
			args := make([]llvm.Value, len(call.Args))
			for i, a := range call.Args {
				args[i] = self.codegenExpr(a, true)
			}
			return self.builder.CreateCall(f, args, "")
		case *hir.MethodCall:
			f := self.vars[call.Method]
			args := make([]llvm.Value, len(call.Args)+1)
			if call.Self.Type().IsPtr() {
				args[0] = self.codegenExpr(call.Self, true)
			} else if !call.Self.Immediate() {
				args[0] = self.codegenExpr(call.Self, false)
			} else {
				selfArg := self.codegenExpr(call.Self, true)
				args[0] = self.builder.CreateAlloca(selfArg.Type(), "")
				self.builder.CreateStore(selfArg, args[0])
			}
			for i, a := range call.Args {
				args[i+1] = self.codegenExpr(a, true)
			}
			v := self.builder.CreateCall(f, args, "")
			if call.Method.NoReturn {
				self.builder.CreateUnreachable()
			}
			return v
		default:
			panic("unreachable")
		}
	case hir.Unary:
		switch unary := expr.(type) {
		case *hir.Not:
			left := self.codegenExpr(unary.Value, true)
			return self.builder.CreateXor(left, llvm.ConstInt(left.Type(), 1, true), "")
		case *hir.GetPointer:
			return self.codegenExpr(unary.Value, false)
		case *hir.GetValue:
			value := self.codegenExpr(unary.Value, true)
			if getValue {
				value = self.builder.CreateLoad(value, "")
			}
			return value
		default:
			panic("unreachable")
		}
	case hir.Index:
		switch index := expr.(type) {
		case *hir.ArrayIndex:
			from, value := self.codegenExpr(index.From, false), self.codegenExpr(index.Index, true)
			return self.createArrayIndex(from, value, getValue)
		case *hir.PointerIndex:
			from, value := self.codegenExpr(index.From, true), self.codegenExpr(index.Index, true)
			return self.createPointerIndex(from, value, getValue)
		case *hir.TupleIndex:
			from := self.codegenExpr(index.From, false)
			return self.createStructIndex(from, index.Index, getValue)
		default:
			panic("unreachable")
		}
	case *hir.Ternary:
		cond := self.builder.CreateIntCast(self.codegenExpr(expr.Cond, true), self.ctx.Int1Type(), "")
		tb, fb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(
			self.function,
			"",
		), llvm.AddBasicBlock(self.function, "")
		self.builder.CreateCondBr(cond, tb, fb)

		self.builder.SetInsertPointAtEnd(tb)
		tv := self.codegenExpr(expr.True, getValue)
		self.builder.CreateBr(eb)

		self.builder.SetInsertPointAtEnd(fb)
		fv := self.codegenExpr(expr.False, getValue)
		self.builder.CreateBr(eb)

		self.builder.SetInsertPointAtEnd(eb)
		phi := self.builder.CreatePHI(tv.Type(), "")
		phi.AddIncoming([]llvm.Value{tv, fv}, []llvm.BasicBlock{tb, fb})
		return phi
	case *hir.Array:
		isConst := true
		elems := make([]llvm.Value, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.codegenExpr(e, true)
			if !elems[i].IsConstant() {
				isConst = false
			}
		}
		if isConst {
			return llvm.ConstArray(self.codegenType(expr.Type()), elems)
		} else {
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type()), "")
			for i, e := range elems {
				index := self.createArrayIndex(tmp, llvm.ConstInt(t_size, uint64(i), false), false)
				self.builder.CreateStore(e, index)
			}
			return self.builder.CreateLoad(tmp, "")
		}
	case *hir.Tuple:
		isConst := true
		elems := make([]llvm.Value, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.codegenExpr(e, true)
			if !elems[i].IsConstant() {
				isConst = false
			}
		}
		if isConst {
			return self.ctx.ConstStruct(elems, false)
		} else {
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type()), "")
			for i, e := range elems {
				index := self.createStructIndex(tmp, uint(i), false)
				self.builder.CreateStore(e, index)
			}
			return self.builder.CreateLoad(tmp, "")
		}
	case *hir.Struct:
		isConst := true
		elems := make([]llvm.Value, len(expr.Fields))
		for i, e := range expr.Fields {
			elems[i] = self.codegenExpr(e, true)
			if !elems[i].IsConstant() {
				isConst = false
			}
		}
		if isConst {
			return self.ctx.ConstStruct(elems, false)
		} else {
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type()), "")
			for i, e := range elems {
				index := self.createStructIndex(tmp, uint(i), false)
				self.builder.CreateStore(e, index)
			}
			return self.builder.CreateLoad(tmp, "")
		}
	case *hir.GetField:
		f := self.codegenExpr(expr.From, false)
		index := expr.GetFieldIndex()
		return self.createStructIndex(f, index, getValue)
	case hir.Covert:
		switch covert := expr.(type) {
		case *hir.WrapCovert:
			return self.codegenExpr(covert.From, true)
		case *hir.Int2Int:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			return self.builder.CreateIntCast(from, to, "")
		case *hir.Float2Float:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			return self.builder.CreateFPCast(from, to, "")
		case *hir.Int2Float:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			switch {
			case covert.From.Type().IsSint():
				return self.builder.CreateSIToFP(from, to, "")
			case covert.From.Type().IsUint():
				return self.builder.CreateUIToFP(from, to, "")
			default:
				panic("unreachable")
			}
		case *hir.Float2Int:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			switch {
			case covert.Type().IsSint():
				return self.builder.CreateFPToSI(from, to, "")
			case covert.Type().IsUint():
				return self.builder.CreateFPToUI(from, to, "")
			default:
				panic("unreachable")
			}
		case *hir.Usize2Ptr:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			return self.builder.CreateIntToPtr(from, to, "")
		case *hir.Ptr2Usize:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			return self.builder.CreatePtrToInt(from, to, "")
		case *hir.Ptr2Ptr:
			from := self.codegenExpr(covert.From, true)
			to := self.codegenType(covert.Type())
			return self.builder.CreatePointerCast(from, to, "")
		default:
			panic("unreachable")
		}
	case *hir.Alloc:
		size := self.codegenExpr(expr.Size, true)
		ptr := self.builder.CreateArrayAlloca(self.ctx.Int8Type(), size, "")
		return self.builder.CreatePointerCast(ptr, llvm.PointerType(t_size, 0), "")
	default:
		panic("unreachable")
	}
}

// 常量表达式
func (self *CodeGenerator) codegenConstantExpr(mean hir.Expr) llvm.Value {
	switch expr := mean.(type) {
	case *hir.Integer:
		value := *(*uint64)(unsafe.Pointer(&expr.Value))
		return llvm.ConstInt(self.codegenType(expr.Typ), value, expr.Typ.IsSint())
	case *hir.Float:
		return llvm.ConstFloat(self.codegenType(expr.Typ), expr.Value)
	case *hir.Boolean:
		return stlutil.Ternary(expr.Value, v_true, v_false)
	case *hir.EmptyFunc, *hir.EmptyPtr:
		return llvm.ConstPointerNull(self.codegenType(expr.Type()))
	case *hir.EmptyArray, *hir.EmptyTuple, *hir.EmptyStruct:
		return llvm.ConstAggregateZero(self.codegenType(expr.Type()))
	case *hir.Array:
		elems := make([]llvm.Value, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.codegenConstantExpr(e)
		}
		return llvm.ConstArray(self.codegenType(expr.Typ), elems)
	case *hir.Tuple:
		elems := make([]llvm.Value, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.codegenConstantExpr(e)
		}
		return self.ctx.ConstStruct(elems, false)
	case *hir.Struct:
		elems := make([]llvm.Value, len(expr.Fields))
		for i, e := range expr.Fields {
			elems[i] = self.codegenConstantExpr(e)
		}
		return self.ctx.ConstStruct(elems, false)
	case *hir.String:
		v, ok := self.cstringPool[expr.Value]
		if !ok {
			init := llvm.ConstString(expr.Value, true)
			vv := llvm.AddGlobal(self.module, init.Type(), "")
			vv.SetGlobalConstant(true)
			vv.SetLinkage(llvm.PrivateLinkage)
			vv.SetInitializer(init)
			v = llvm.AddGlobal(self.module, self.codegenType(expr.Type()), "")
			v.SetGlobalConstant(true)
			v.SetLinkage(llvm.PrivateLinkage)
			v.SetInitializer(llvm.ConstPointerCast(vv, v.Type().ElementType()))
			self.cstringPool[expr.Value] = v
		}
		return self.builder.CreateLoad(v, "")
	default:
		panic("unreachable")
	}
}

// 比较
func (self *CodeGenerator) equal(left, right llvm.Value) llvm.Value {
	switch left.Type().TypeKind() {
	case llvm.IntegerTypeKind, llvm.PointerTypeKind, llvm.FunctionTypeKind:
		return self.builder.CreateICmp(llvm.IntEQ, left, right, "")
	case llvm.FloatTypeKind:
		return self.builder.CreateFCmp(llvm.FloatOEQ, left, right, "")
	case llvm.ArrayTypeKind:
		if left.Type().ArrayLength() == 0 {
			return llvm.ConstInt(self.ctx.Int8Type(), 1, true)
		}
		i := self.builder.CreateAlloca(self.codegenType(hir.NewTypeUsize()), "")
		self.builder.CreateStore(llvm.ConstInt(i.Type().ElementType(), 0, false), i)
		cb := llvm.AddBasicBlock(self.function, "")
		self.builder.CreateBr(cb)

		self.builder.SetInsertPointAtEnd(cb)
		iv := self.builder.CreateLoad(i, "")
		lb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
		lt := self.builder.CreateICmp(
			llvm.IntULT,
			iv,
			llvm.ConstInt(iv.Type(), uint64(left.Type().ArrayLength()), false),
			"",
		)
		self.builder.CreateCondBr(lt, lb, eb)

		self.builder.SetInsertPointAtEnd(lb)
		l, r := self.createArrayIndex(left, iv, true), self.createArrayIndex(right, iv, true)
		self.builder.CreateStore(self.builder.CreateNUWAdd(iv, llvm.ConstInt(iv.Type(), 1, false), ""), i)
		self.builder.CreateCondBr(self.equal(l, r), cb, eb)
		lb = self.builder.GetInsertBlock()

		self.builder.SetInsertPointAtEnd(eb)
		phi := self.builder.CreatePHI(t_bool, "")
		phi.AddIncoming([]llvm.Value{v_true, v_false}, []llvm.BasicBlock{cb, lb})
		return phi
	case llvm.StructTypeKind:
		elemCount := left.Type().StructElementTypesCount()
		if elemCount == 0 {
			return llvm.ConstInt(self.ctx.Int8Type(), 1, true)
		}
		blocks := make([]llvm.BasicBlock, elemCount)
		values := make([]llvm.Value, elemCount)
		eb := llvm.AddBasicBlock(self.function, "")
		for i := range left.Type().StructElementTypes() {
			l, r := self.createStructIndex(left, uint(i), true), self.createStructIndex(right, uint(i), true)
			v := self.equal(l, r)
			blocks[i], values[i] = self.builder.GetInsertBlock(), v
			if i < elemCount-1 {
				nb := llvm.AddBasicBlock(self.function, "")
				self.builder.CreateCondBr(v, nb, eb)
				self.builder.SetInsertPointAtEnd(nb)
			} else {
				self.builder.CreateBr(eb)
				self.builder.SetInsertPointAtEnd(eb)
			}
		}
		phi := self.builder.CreatePHI(self.ctx.Int1Type(), "")
		phi.AddIncoming(values, blocks)
		return phi
	default:
		panic("")
	}
}
