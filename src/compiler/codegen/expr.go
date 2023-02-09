package codegen

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
	stlutil "github.com/kkkunny/stl/util"
)

// 表达式
func (self *CodeGenerator) codegenExpr(mean hir.Expr, getValue bool) llvm.Value {
	switch expr := mean.(type) {
	case *hir.Null, *hir.Integer, *hir.Float, *hir.Boolean, *hir.String, *hir.EmptyStruct, *hir.EmptyArray, *hir.EmptyTuple:
		return self.codegenConstantExpr(mean)
	case *hir.Binary:
		switch expr.Opera {
		case "+":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNSWAdd(l, r, "")
			} else if hir.IsUintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNUWAdd(l, r, "")
			} else {
				return self.builder.CreateFAdd(l, r, "")
			}
		case "-":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNSWSub(l, r, "")
			} else if hir.IsUintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNUWSub(l, r, "")
			} else {
				return self.builder.CreateFSub(l, r, "")
			}
		case "*":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNSWMul(l, r, "")
			} else if hir.IsUintTypeAndSon(expr.GetType()) {
				return self.builder.CreateNUWMul(l, r, "")
			} else {
				return self.builder.CreateFMul(l, r, "")
			}
		case "/":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateSDiv(l, r, "")
			} else if hir.IsUintTypeAndSon(expr.GetType()) {
				return self.builder.CreateUDiv(l, r, "")
			} else {
				return self.builder.CreateFDiv(l, r, "")
			}
		case "%":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateSRem(l, r, "")
			} else if hir.IsUintTypeAndSon(expr.GetType()) {
				return self.builder.CreateURem(l, r, "")
			} else {
				return self.builder.CreateFRem(l, r, "")
			}
		case "&":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			return self.builder.CreateAnd(l, r, "")
		case "|":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			return self.builder.CreateOr(l, r, "")
		case "^":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			return self.builder.CreateXor(l, r, "")
		case "<<":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			return self.builder.CreateShl(l, r, "")
		case ">>":
			l, r := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
			if hir.IsSintTypeAndSon(expr.GetType()) {
				return self.builder.CreateAShr(l, r, "")
			} else {
				return self.builder.CreateLShr(l, r, "")
			}
		case "&&":
			nb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
			self.builder.CreateCondBr(self.builder.CreateIntCast(self.codegenExpr(expr.Left, true), self.ctx.Int1Type(), ""), nb, eb)
			pb := self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(nb)
			nv := self.builder.CreateIntCast(self.codegenExpr(expr.Right, true), self.ctx.Int1Type(), "")
			self.builder.CreateBr(eb)
			nb = self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(eb)
			phi := self.builder.CreatePHI(nv.Type(), "")
			phi.AddIncoming([]llvm.Value{llvm.ConstInt(self.ctx.Int1Type(), 0, true), nv}, []llvm.BasicBlock{pb, nb})
			return self.builder.CreateIntCast(phi, t_bool, "")
		case "||":
			nb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
			self.builder.CreateCondBr(self.builder.CreateIntCast(self.codegenExpr(expr.Left, true), self.ctx.Int1Type(), ""), eb, nb)
			pb := self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(nb)
			nv := self.builder.CreateIntCast(self.codegenExpr(expr.Right, true), self.ctx.Int1Type(), "")
			self.builder.CreateBr(eb)
			nb = self.builder.GetInsertBlock()

			self.builder.SetInsertPointAtEnd(eb)
			phi := self.builder.CreatePHI(nv.Type(), "")
			phi.AddIncoming([]llvm.Value{llvm.ConstInt(self.ctx.Int1Type(), 1, true), nv}, []llvm.BasicBlock{pb, nb})
			return self.builder.CreateIntCast(phi, t_bool, "")
		default:
			panic("")
		}
	case *hir.Variable:
		v := self.vars[expr]
		if getValue {
			v = self.builder.CreateLoad(v, "")
		}
		return v
	case *hir.Function:
		return self.vars[expr]
	case *hir.Method:
		return self.vars[expr.Func]
	case *hir.FuncCall:
		f := self.codegenExpr(expr.Func, true)
		args := make([]llvm.Value, len(expr.Args))
		for i, a := range expr.Args {
			args[i] = self.codegenExpr(a, true)
		}
		call := self.builder.CreateCall(f, args, "")
		return call
	case *hir.MethodCall:
		f := self.codegenExpr(expr.Method.Func, true)
		args := make([]llvm.Value, len(expr.Args)+1)
		if hir.IsPtrType(expr.Method.Self.GetType()) {
			args[0] = self.codegenExpr(expr.Method.Self, true)
		} else if expr.Method.Self.GetMut() {
			args[0] = self.codegenExpr(expr.Method.Self, false)
		} else {
			selfArg := self.codegenExpr(expr.Method.Self, true)
			args[0] = self.builder.CreateAlloca(selfArg.Type(), "")
			self.builder.CreateStore(selfArg, args[0])
		}
		for i, a := range expr.Args {
			args[i+1] = self.codegenExpr(a, true)
		}
		call := self.builder.CreateCall(f, args, "")
		if expr.Method.Func.NoReturn {
			self.builder.CreateUnreachable()
		}
		return call
	case *hir.Param:
		v := self.vars[expr]
		if getValue {
			v = self.builder.CreateLoad(v, "")
		}
		return v
	case *hir.Assign:
		switch expr.Opera {
		case "=":
			left, right := self.codegenExpr(expr.Left, false), self.codegenExpr(expr.Right, true)
			self.builder.CreateStore(right, left)
			return llvm.Value{}
		default:
			return self.codegenExpr(&hir.Assign{
				Opera: "=",
				Left:  expr.Left,
				Right: &hir.Binary{
					Opera: expr.Opera[:len(expr.Opera)-1],
					Left:  expr.Left,
					Right: expr.Right,
				},
			}, true)
		}
	case *hir.Equal:
		left, right := self.codegenExpr(expr.Left, true), self.codegenExpr(expr.Right, true)
		var v llvm.Value
		switch expr.Opera {
		case "==":
			v = self.equal(left, right)
		case "!=":
			left = self.equal(left, right)
			v = self.builder.CreateXor(left, llvm.ConstInt(left.Type(), 1, true), "")
		case "<":
			if hir.IsSintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntSLT, left, right, "")
			} else if hir.IsUintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntULT, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOLT, left, right, "")
			}
		case "<=":
			if hir.IsSintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntSLE, left, right, "")
			} else if hir.IsUintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntULE, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOLE, left, right, "")
			}
		case ">":
			if hir.IsSintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntSGT, left, right, "")
			} else if hir.IsUintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntUGT, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOGT, left, right, "")
			}
		case ">=":
			if hir.IsSintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntSGE, left, right, "")
			} else if hir.IsUintTypeAndSon(expr.Left.GetType()) {
				v = self.builder.CreateICmp(llvm.IntUGE, left, right, "")
			} else {
				v = self.builder.CreateFCmp(llvm.FloatOGE, left, right, "")
			}
		default:
			panic(fmt.Sprintf("unknown equal: %+v", expr))
		}
		return self.builder.CreateIntCast(v, t_bool, "")
	case *hir.Unary:
		switch expr.Opera {
		case "!":
			left := self.codegenExpr(expr.Value, true)
			return self.builder.CreateXor(left, llvm.ConstInt(left.Type(), 1, true), "")
		case "&":
			return self.codegenExpr(expr.Value, false)
		case "*":
			value := self.codegenExpr(expr.Value, true)
			if getValue {
				value = self.builder.CreateLoad(value, "")
			}
			return value
		default:
			panic("")
		}
	case *hir.Index:
		fromType := expr.From.GetType()
		switch {
		case hir.IsArrayTypeAndSon(fromType):
			from, index := self.codegenExpr(expr.From, false), self.codegenExpr(expr.Index, true)
			return self.createArrayIndex(from, index, getValue)
		case hir.IsPtrTypeAndSon(fromType):
			from, index := self.codegenExpr(expr.From, true), self.codegenExpr(expr.Index, true)
			return self.createPointerIndex(from, index, getValue)
		case hir.IsTupleTypeAndSon(fromType):
			from := self.codegenExpr(expr.From, false)
			index := expr.Index.(*hir.Integer).Value
			return self.createStructIndex(from, uint(index), getValue)
		default:
			panic("")
		}
	case *hir.Select:
		cond := self.builder.CreateIntCast(self.codegenExpr(expr.Cond, true), self.ctx.Int1Type(), "")
		tb, fb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
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
			return llvm.ConstArray(self.codegenType(expr.Type), elems)
		} else {
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type), "")
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
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type), "")
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
			tmp := self.builder.CreateAlloca(self.codegenType(expr.Type), "")
			for i, e := range elems {
				index := self.createStructIndex(tmp, uint(i), false)
				self.builder.CreateStore(e, index)
			}
			return self.builder.CreateLoad(tmp, "")
		}
	case *hir.GetField:
		f := self.codegenExpr(expr.From, false)
		var index uint
		for iter := hir.GetBaseType(expr.From.GetType()).(*hir.TypeStruct).Fields.Begin(); iter.HasValue(); iter.Next() {
			if iter.Key() == expr.Index {
				break
			}
			index++
		}
		return self.createStructIndex(f, index, getValue)
	case *hir.Covert:
		from := self.codegenExpr(expr.From, true)
		meanFt, meanTo := expr.From.GetType(), expr.To
		to := self.codegenType(expr.GetType())
		switch {
		case hir.GetDepthBaseType(meanFt).Equal(hir.GetDepthBaseType(meanTo)):
			return from
		case hir.IsIntTypeAndSon(meanFt) && hir.IsIntTypeAndSon(meanTo):
			return self.builder.CreateIntCast(from, to, "")
		case hir.IsFloatTypeAndSon(meanFt) && hir.IsFloatTypeAndSon(meanTo):
			return self.builder.CreateFPCast(from, to, "")
		case hir.IsSintTypeAndSon(meanFt) && hir.IsFloatTypeAndSon(meanTo):
			return self.builder.CreateSIToFP(from, to, "")
		case hir.IsUintTypeAndSon(meanFt) && hir.IsFloatTypeAndSon(meanTo):
			return self.builder.CreateUIToFP(from, to, "")
		case hir.IsFloatTypeAndSon(meanFt) && hir.IsSintTypeAndSon(meanTo):
			return self.builder.CreateFPToSI(from, to, "")
		case hir.IsFloatTypeAndSon(meanFt) && hir.IsUintTypeAndSon(meanTo):
			return self.builder.CreateFPToUI(from, to, "")
		case hir.GetBaseType(meanFt).Equal(hir.Usize) && (hir.IsPtrTypeAndSon(meanTo) || hir.IsFuncTypeAndSon(meanTo)):
			return self.builder.CreateIntToPtr(from, to, "")
		case (hir.IsPtrTypeAndSon(meanFt) || hir.IsFuncTypeAndSon(meanFt)) && hir.GetBaseType(meanTo).Equal(hir.Usize):
			return self.builder.CreatePtrToInt(from, to, "")
		case (hir.IsPtrTypeAndSon(meanFt) || hir.IsFuncTypeAndSon(meanFt)) && (hir.IsPtrTypeAndSon(meanTo) || hir.IsFuncTypeAndSon(meanTo)):
			return self.builder.CreatePointerCast(from, to, "")
		case hir.IsPtrTypeAndSon(meanFt) && hir.IsTypedef(hir.GetBaseType(meanFt).(*hir.TypePtr).Elem) && hir.IsInterfaceTypeAndSon(meanTo) && hir.GetBaseType(meanFt).(*hir.TypePtr).Elem.(*hir.Typedef).IsImpl(hir.GetBaseType(meanTo).(*hir.TypeInterface)):
			ft := hir.GetBaseType(meanFt).(*hir.TypePtr).Elem.(*hir.Typedef)
			it := hir.GetBaseType(meanTo).(*hir.TypeInterface)
			toParams := to.StructElementTypes()
			alloca := self.builder.CreateAlloca(to, "")

			index := self.createStructIndex(alloca, 0, false)
			self.builder.CreateStore(self.codegenExpr(&hir.String{
				Type:  hir.NewPtrType(hir.I8),
				Value: ft.String(),
			}, false), index)
			index = self.createStructIndex(alloca, 1, false)
			self.builder.CreateStore(self.builder.CreatePointerCast(from, toParams[1], ""), index)
			for iter := it.Fields.Begin(); iter.HasValue(); iter.Next() {
				index = self.createStructIndex(alloca, uint(iter.Index()+2), false)
				f := self.codegenExpr(ft.Methods[iter.Key()], true)
				ptr := self.builder.CreatePointerCast(f, toParams[iter.Index()+2], "")
				self.builder.CreateStore(ptr, index)
			}

			return self.builder.CreateLoad(alloca, "")
		default:
			panic("")
		}
	case *hir.GlobalVariable:
		v := self.vars[expr]
		if getValue {
			v = self.builder.CreateLoad(v, "")
		}
		return v
	case *hir.Alloc:
		size := self.codegenExpr(expr.Size, true)
		ptr := self.builder.CreateArrayAlloca(self.ctx.Int8Type(), size, "")
		return self.builder.CreatePointerCast(ptr, llvm.PointerType(t_size, 0), "")
	case *hir.GetInterfaceField:
		f := self.codegenExpr(expr.From, false)
		var index uint = 2
		for iter := hir.GetBaseType(expr.From.GetType()).(*hir.TypeInterface).Fields.Begin(); iter.HasValue(); iter.Next() {
			if iter.Key() == expr.Index {
				break
			}
			index++
		}
		return self.createStructIndex(f, index, true)
	case *hir.InterfaceFieldCall:
		f := self.codegenExpr(expr.Field, true)

		args := make([]llvm.Value, len(expr.Args)+1)
		args[0] = self.createStructIndex(self.codegenExpr(expr.Field.From, true), 1, true)
		for i, a := range expr.Args {
			args[i+1] = self.codegenExpr(a, true)
		}

		return self.builder.CreateCall(f, args, "")
	default:
		panic("")
	}
}

// 常量表达式
func (self *CodeGenerator) codegenConstantExpr(mean hir.Expr) llvm.Value {
	switch expr := mean.(type) {
	case *hir.Null:
		return llvm.ConstPointerNull(self.codegenType(expr.Type))
	case *hir.Integer:
		value := *(*uint64)(unsafe.Pointer(&expr.Value))
		return llvm.ConstInt(self.codegenType(expr.Type), value, hir.IsSintType(expr.Type))
	case *hir.Float:
		return llvm.ConstFloat(self.codegenType(expr.Type), expr.Value)
	case *hir.Boolean:
		return stlutil.Ternary(expr.Value, v_true, v_false)
	case *hir.EmptyArray, *hir.EmptyTuple, *hir.EmptyStruct:
		return llvm.ConstAggregateZero(self.codegenType(expr.GetType()))
	case *hir.Array:
		elems := make([]llvm.Value, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.codegenConstantExpr(e)
		}
		return llvm.ConstArray(self.codegenType(expr.Type), elems)
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
			v = llvm.AddGlobal(self.module, self.codegenType(expr.GetType()), "")
			v.SetGlobalConstant(true)
			v.SetLinkage(llvm.PrivateLinkage)
			v.SetInitializer(llvm.ConstPointerCast(vv, v.Type().ElementType()))
			self.cstringPool[expr.Value] = v
		}
		return self.builder.CreateLoad(v, "")
	default:
		panic("")
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
		i := self.builder.CreateAlloca(self.codegenType(hir.Usize), "")
		self.builder.CreateStore(llvm.ConstInt(i.Type().ElementType(), 0, false), i)
		cb := llvm.AddBasicBlock(self.function, "")
		self.builder.CreateBr(cb)

		self.builder.SetInsertPointAtEnd(cb)
		iv := self.builder.CreateLoad(i, "")
		lb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
		lt := self.builder.CreateICmp(llvm.IntULT, iv, llvm.ConstInt(iv.Type(), uint64(left.Type().ArrayLength()), false), "")
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
