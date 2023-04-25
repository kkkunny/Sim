package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/llvm"
)

// 代码块
func (self *CodeGenerator) codegenBlock(ir *mir.Block) {
	self.builder.SetInsertPointAtEnd(self.blocks[ir])
	for cursor := ir.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		self.codegenInst(cursor.Value())
	}
}

// 语句
func (self *CodeGenerator) codegenInst(ir mir.Inst) {
	var value llvm.Value

	switch inst := ir.(type) {
	case *mir.Alloc:
		value = self.builder.CreateAlloca(self.codegenType(inst.Type), "")
		value.SetAlignment(int(inst.Type.Align()))
	case *mir.Load:
		from := self.codegenValue(inst.Value)
		value = self.builder.CreateLoad(from, "")
		value.SetAlignment(int(inst.Value.GetType().GetPtr().Align()))
	case *mir.Store:
		from, to := self.codegenValue(inst.From), self.codegenValue(inst.To)
		store := self.builder.CreateStore(from, to)
		store.SetAlignment(int(inst.From.GetType().Align()))
	case *mir.Unreachable:
		self.builder.CreateUnreachable()
	case *mir.Return:
		if inst.Value == nil {
			self.builder.CreateRetVoid()
		} else {
			v := self.codegenValue(inst.Value)
			self.builder.CreateRet(v)
		}
	case *mir.Jmp:
		self.builder.CreateBr(self.blocks[inst.Dst])
	case *mir.CondJmp:
		cond := self.codegenValue(inst.Cond)
		self.builder.CreateCondBr(cond, self.blocks[inst.True], self.blocks[inst.False])
	case *mir.Eq:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsInteger() || leftMirType.IsBool() || leftMirType.IsPtr() || leftMirType.IsFunc():
			value = self.builder.CreateICmp(llvm.IntEQ, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatOEQ, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Ne:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsInteger() || leftMirType.IsBool() || leftMirType.IsPtr() || leftMirType.IsFunc():
			value = self.builder.CreateICmp(llvm.IntNE, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatUNE, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Lt:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateICmp(llvm.IntSLT, l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateICmp(llvm.IntULT, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatOLT, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Le:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateICmp(llvm.IntSLE, l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateICmp(llvm.IntULE, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatOLE, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Gt:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateICmp(llvm.IntSGT, l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateICmp(llvm.IntUGT, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatOGT, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Ge:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateICmp(llvm.IntSGE, l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateICmp(llvm.IntUGE, l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFCmp(llvm.FloatOGE, l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.ArrayIndex:
		from, index := self.codegenValue(inst.From), self.codegenValue(inst.Index)
		if inst.IsPtr() {
			value = self.builder.CreateInBoundsGEP(from, []llvm.Value{llvm.ConstInt(index.Type(), 0, false), index}, "")
		} else {
			value = self.builder.CreateExtractElement(from, index, "")
		}
	case *mir.StructIndex:
		from := self.codegenValue(inst.From)
		if inst.IsPtr() {
			value = self.builder.CreateStructGEP(from, int(inst.Index), "")
		} else {
			value = self.builder.CreateExtractValue(from, int(inst.Index), "")
		}
	case *mir.PointerIndex:
		from, index := self.codegenValue(inst.From), self.codegenValue(inst.Index)
		value = self.builder.CreateInBoundsGEP(from, []llvm.Value{index}, "")
	case *mir.UnwrapUnion:
		v, t := self.codegenValue(inst.Value), self.codegenType(inst.To)
		if inst.IsPtr() {
			value = self.builder.CreateBitCast(v, llvm.PointerType(t, 0), "")
		} else {
			tmp := self.builder.CreateAlloca(v.Type(), "")
			tmp.SetAlignment(int(inst.Value.GetType().Align()))
			store := self.builder.CreateStore(v, tmp)
			store.SetAlignment(int(inst.Value.GetType().Align()))
			value = self.builder.CreateBitCast(tmp, llvm.PointerType(t, 0), "")
			value = self.builder.CreateLoad(value, "")
			value.SetAlignment(int(inst.To.Align()))
			ptr := self.builder.CreateBitCast(tmp, llvm.PointerType(self.ctx.Int8Type(), 0), "")
			self.CreateCallLLVMLifetimeEnd(
				llvm.ConstInt(
					self.ctx.Int64Type(),
					uint64(inst.Value.GetType().Size()),
					true,
				), ptr,
			)
		}
	case *mir.Add:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateNSWAdd(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateNUWAdd(l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFAdd(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Sub:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateNSWSub(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateNUWSub(l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFSub(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Mul:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateNSWMul(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateNUWMul(l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFMul(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Div:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateSDiv(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateUDiv(l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFDiv(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Mod:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateSRem(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateURem(l, r, "")
		case leftMirType.IsFloat():
			value = self.builder.CreateFRem(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.And:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateAnd(l, r, "")
	case *mir.Or:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateOr(l, r, "")
	case *mir.Xor:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateXor(l, r, "")
	case *mir.Shl:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateShl(l, r, "")
	case *mir.Shr:
		leftMirType := inst.Left.GetType()
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		switch {
		case leftMirType.IsSint():
			value = self.builder.CreateAShr(l, r, "")
		case leftMirType.IsUint():
			value = self.builder.CreateLShr(l, r, "")
		default:
			panic("unreachable")
		}
	case *mir.Phi:
		vs, bs := make([]llvm.Value, len(inst.Values)), make([]llvm.BasicBlock, len(inst.Froms))
		for i, f := range inst.Froms {
			bs[i] = self.blocks[f]
		}
		for i, v := range inst.Values {
			vs[i] = self.codegenValue(v)
		}
		value = self.builder.CreatePHI(self.codegenType(inst.GetType()), "")
		value.AddIncoming(vs, bs)
	case *mir.Sint2Sint:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = v
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateSExt(v, t, "")
		} else {
			value = self.builder.CreateTrunc(v, t, "")
		}
	case *mir.Sint2Uint:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = v
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateSExt(v, t, "")
		} else {
			value = self.builder.CreateTrunc(v, t, "")
		}
	case *mir.Sint2Float:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreateSIToFP(v, t, "")
	case *mir.Sint2Ptr:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = self.builder.CreateIntToPtr(v, t, "")
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateSExt(v, self.targetData.IntPtrType(), "")
			value = self.builder.CreateIntToPtr(value, t, "")
		} else {
			value = self.builder.CreateTrunc(v, self.targetData.IntPtrType(), "")
			value = self.builder.CreateIntToPtr(value, t, "")
		}
	case *mir.Uint2Uint:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = v
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateZExt(v, t, "")
		} else {
			value = self.builder.CreateTrunc(v, t, "")
		}
	case *mir.Uint2Sint:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = v
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateZExt(v, t, "")
		} else {
			value = self.builder.CreateTrunc(v, t, "")
		}
	case *mir.Uint2Float:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreateUIToFP(v, t, "")
	case *mir.Uint2Ptr:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = self.builder.CreateIntToPtr(v, t, "")
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateZExt(v, self.targetData.IntPtrType(), "")
			value = self.builder.CreateIntToPtr(value, t, "")
		} else {
			value = self.builder.CreateTrunc(v, self.targetData.IntPtrType(), "")
			value = self.builder.CreateIntToPtr(value, t, "")
		}
	case *mir.Float2Sint:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreateFPToSI(v, t, "")
	case *mir.Float2Uint:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreateFPToUI(v, t, "")
	case *mir.Float2Float:
		fromMirType := inst.From.GetType()
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		if fromMirType.Size() == inst.To.Size() {
			value = v
		} else if fromMirType.Size() < inst.To.Size() {
			value = self.builder.CreateFPExt(v, t, "")
		} else {
			value = self.builder.CreateFPTrunc(v, t, "")
		}
	case *mir.Ptr2Ptr:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreateBitCast(v, t, "")
	case *mir.Ptr2Sint:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreatePtrToInt(v, t, "")
	case *mir.Ptr2Uint:
		v, t := self.codegenValue(inst.From), self.codegenType(inst.To)
		value = self.builder.CreatePtrToInt(v, t, "")
	case *mir.Select:
		c, t, f := self.codegenValue(inst.Cond), self.codegenValue(inst.True), self.codegenValue(inst.False)
		value = self.builder.CreateSelect(c, t, f, "")
	case *mir.Not:
		v := self.codegenValue(inst.Value)
		value = self.builder.CreateNot(v, "")
	case *mir.LogicNot:
		v := self.codegenValue(inst.Value)
		value = self.builder.CreateICmp(llvm.IntEQ, v, llvm.ConstInt(v.Type(), 0, true), "")
	case *mir.LogicAnd:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateAnd(l, r, "")
	case *mir.LogicOr:
		l, r := self.codegenValue(inst.Left), self.codegenValue(inst.Right)
		value = self.builder.CreateOr(l, r, "")
	case *mir.Call:
		f := self.codegenValue(inst.Func)
		args := make([]llvm.Value, len(inst.Args))
		for i, a := range inst.Args {
			args[i] = self.codegenValue(a)
		}
		value = self.builder.CreateCall(f, args, "")
	case *mir.WrapUnion:
		v, t := self.codegenValue(inst.Value), self.codegenType(inst.To)
		tmp := self.builder.CreateAlloca(t, "")
		tmp.SetAlignment(int(inst.To.Align()))
		ptr := self.builder.CreateBitCast(tmp, llvm.PointerType(v.Type(), 0), "")
		store := self.builder.CreateStore(v, ptr)
		store.SetAlignment(int(inst.Value.GetType().Align()))
		value = self.builder.CreateLoad(tmp, "")
		value.SetAlignment(int(inst.To.Align()))
		ptr = self.builder.CreateBitCast(tmp, llvm.PointerType(self.ctx.Int8Type(), 0), "")
		self.CreateCallLLVMLifetimeEnd(llvm.ConstInt(self.ctx.Int64Type(), uint64(inst.To.Size()), true), ptr)
	default:
		panic("unreachable")
	}

	if v, ok := ir.(mir.Value); ok {
		self.vars[v] = value
	}
}
