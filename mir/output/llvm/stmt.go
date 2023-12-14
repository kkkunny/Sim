package llvm

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

func (self *LLVMOutputer) codegenStmt(ir mir.Stmt){
	switch stmt := ir.(type) {
	case *mir.Store:
		inst := self.builder.CreateStore(self.codegenValue(stmt.From()), self.codegenValue(stmt.To()))
		inst.SetAlign(uint32(stmt.From().Type().Align()))
	case *mir.Return:
		valueIr, ok := stmt.Value()
		if !ok{
			self.builder.CreateRet(nil)
			return
		}
		value := self.codegenValue(valueIr)

		fir := stmt.Belong().Belong()
		f := self.values.Get(fir).(llvm.Function)
		switch {
		case self.target.IsWindows() && self.codegenType(fir.Type().(mir.FuncType).Ret()).String() != f.FunctionType().ReturnType().String():
			self.builder.CreateStore(value, f.GetParam(0))
			self.builder.CreateRet(nil)
		default:
			self.builder.CreateRet(&value)
		}
	case *mir.UnCondJump:
		self.builder.CreateBr(self.blocks.Get(stmt.To()))
	case *mir.CondJump:
		self.builder.CreateCondBr(self.codegenValue(stmt.Cond()), self.blocks.Get(stmt.TrueBlock()), self.blocks.Get(stmt.FalseBlock()))
	case mir.StmtValue:
		self.codegenStmtValue(stmt)
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenStmtValue(ir mir.StmtValue)(res llvm.Value){
	defer func() {
		self.values.Set(ir, res)
	}()

	switch sv := ir.(type) {
	case *mir.AllocFromStack:
		inst := self.builder.CreateAlloca("", self.codegenType(sv.ElemType()))
		inst.SetAlign(uint32(sv.ElemType().Align()))
		return inst
	case *mir.AllocFromHeap:
		return self.builder.CreateMalloc("", self.codegenType(sv.ElemType()))
	case *mir.Load:
		inst := self.builder.CreateLoad("", self.codegenType(sv.Type()), self.codegenValue(sv.From()))
		inst.SetAlign(uint32(sv.Type().Align()))
		return inst
	case *mir.And:
		return self.builder.CreateAnd("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
	case *mir.Or:
		return self.builder.CreateOr("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
	case *mir.Xor:
		return self.builder.CreateXor("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
	case *mir.Shl:
		return self.builder.CreateShl("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
	case *mir.Shr:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateAShr("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateLShr("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Not:
		return self.builder.CreateNot("", self.codegenValue(sv.Value()))
	case *mir.Add:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateSAdd("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateUAdd("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.FloatType:
			return self.builder.CreateFAdd("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Sub:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateSSub("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateUSub("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.FloatType:
			return self.builder.CreateFSub("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Mul:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateSMul("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateUMul("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.FloatType:
			return self.builder.CreateFMul("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Div:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateSDiv("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateUDiv("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.FloatType:
			return self.builder.CreateFDiv("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Rem:
		switch sv.Type().(type) {
		case mir.SintType:
			return self.builder.CreateSRem("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.UintType:
			return self.builder.CreateURem("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.FloatType:
			return self.builder.CreateFRem("", self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Cmp:
		switch sv.Kind() {
		case mir.CmpKindEQ:
			switch sv.Left().Type().(type) {
			case mir.IntType:
				return self.builder.CreateIntCmp("", llvm.IntEQ, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatOEQ, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		case mir.CmpKindNE:
			switch sv.Left().Type().(type) {
			case mir.IntType:
				return self.builder.CreateIntCmp("", llvm.IntNE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatUNE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		case mir.CmpKindLT:
			switch sv.Left().Type().(type) {
			case mir.SintType:
				return self.builder.CreateIntCmp("", llvm.IntSLT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.UintType:
				return self.builder.CreateIntCmp("", llvm.IntULT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatOLT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		case mir.CmpKindLE:
			switch sv.Left().Type().(type) {
			case mir.SintType:
				return self.builder.CreateIntCmp("", llvm.IntSLE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.UintType:
				return self.builder.CreateIntCmp("", llvm.IntULE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatOLE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		case mir.CmpKindGT:
			switch sv.Left().Type().(type) {
			case mir.SintType:
				return self.builder.CreateIntCmp("", llvm.IntSGT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.UintType:
				return self.builder.CreateIntCmp("", llvm.IntUGT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatOGT, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		case mir.CmpKindGE:
			switch sv.Left().Type().(type) {
			case mir.SintType:
				return self.builder.CreateIntCmp("", llvm.IntSGE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.UintType:
				return self.builder.CreateIntCmp("", llvm.IntUGE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			case mir.FloatType:
				return self.builder.CreateFloatCmp("", llvm.FloatOGE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
			default:
				panic("unreachable")
			}
		default:
			panic("unreachable")
		}
	case *mir.PtrEqual:
		switch sv.Kind() {
		case mir.PtrEqualKindEQ:
			return self.builder.CreateIntCmp("", llvm.IntEQ, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		case mir.PtrEqualKindNE:
			return self.builder.CreateIntCmp("", llvm.IntNE, self.codegenValue(sv.Left()), self.codegenValue(sv.Right()))
		default:
			panic("unreachable")
		}
	case *mir.Call:
		f := self.codegenValue(sv.Func())
		args := lo.Map(sv.Args(), func(item mir.Value, _ int) llvm.Value {
			return self.codegenValue(item)
		})

		switch {
		case self.target.IsWindows():
			retType := self.codegenType(sv.Type())
			targetType := self.codegenFuncType(sv.Func().Type().(mir.FuncType))
			if retType.String() != targetType.ReturnType().String(){
				ptr := self.builder.CreateAlloca("", retType)
				args = append([]llvm.Value{ptr}, args...)
			}

			for i, a := range args {
				if at := a.Type(); stlbasic.Is[llvm.StructType](at) || stlbasic.Is[llvm.Array](at) {
					ptr := self.builder.CreateAlloca("", at)
					self.builder.CreateStore(a, ptr)
					args[i] = ptr
				}
			}

			var ret llvm.Value = self.builder.CreateCall("", targetType, f, args...)
			if retType.String() != targetType.ReturnType().String(){
				ret = self.builder.CreateLoad("", retType, args[0])
			}
			return ret
		default:
			return self.builder.CreateCall("", self.codegenFuncType(sv.Func().Type().(mir.FuncType)), f, args...)
		}
	case *mir.NumberCovert:
		ftIr, ttIr := sv.From().Type(), sv.Type()
		ftSize, ttSize := ftIr.Size(),ttIr.Size()
		from := self.codegenValue(sv.From())
		tt := self.codegenType(ttIr)
		switch {
		case stlbasic.Is[mir.SintType](ftIr) && stlbasic.Is[mir.IntType](ttIr):
			if ftSize < ttSize{
				return self.builder.CreateSExt("", from, tt.(llvm.IntegerType))
			}else if ftSize > ttSize{
				return self.builder.CreateTrunc("", from, tt.(llvm.IntegerType))
			}else{
				return from
			}
		case stlbasic.Is[mir.UintType](ftIr) && stlbasic.Is[mir.IntType](ttIr):
			if ftSize < ttSize{
				return self.builder.CreateZExt("", from, tt.(llvm.IntegerType))
			}else if ftSize > ttSize{
				return self.builder.CreateTrunc("", from, tt.(llvm.IntegerType))
			}else{
				return from
			}
		case stlbasic.Is[mir.FloatType](ftIr) && stlbasic.Is[mir.FloatType](ttIr):
			if ftSize < ttSize{
				return self.builder.CreateFPExt("", from, tt.(llvm.FloatType))
			}else if ftSize > ttSize{
				return self.builder.CreateFPTrunc("", from, tt.(llvm.FloatType))
			}else{
				return from
			}
		case stlbasic.Is[mir.SintType](ftIr) && stlbasic.Is[mir.FloatType](ttIr):
			return self.builder.CreateSIToFP("", from, tt.(llvm.FloatType))
		case stlbasic.Is[mir.UintType](ftIr) && stlbasic.Is[mir.FloatType](ttIr):
			return self.builder.CreateUIToFP("", from, tt.(llvm.FloatType))
		case stlbasic.Is[mir.FloatType](ftIr) && stlbasic.Is[mir.SintType](ttIr):
			return self.builder.CreateFPToSI("", from, tt.(llvm.IntegerType))
		case stlbasic.Is[mir.FloatType](ftIr) && stlbasic.Is[mir.UintType](ttIr):
			return self.builder.CreateFPToUI("", from, tt.(llvm.IntegerType))
		default:
			panic("unreachable")
		}
	case *mir.PtrToUint:
		return self.builder.CreatePtrToInt("", self.codegenValue(sv.From()), self.codegenIntType(sv.Type().(mir.IntType)))
	case *mir.UintToPtr:
		return self.builder.CreateIntToPtr("", self.codegenValue(sv.From()), self.codegenType(sv.Type()).(llvm.PointerType))
	case *mir.PtrToPtr:
		return self.builder.CreateBitCast("", self.codegenValue(sv.From()), self.codegenType(sv.Type()))
	case *mir.ArrayIndex:
		array := self.codegenValue(sv.Array())
		index := self.codegenValue(sv.Index())
		if sv.IsPtr(){
			at := self.codegenArrayType(sv.Array().Type().(mir.PtrType).Elem().(mir.ArrayType))
			return self.builder.CreateInBoundsGEP("", at, array, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
		}else{
			at := self.codegenArrayType(sv.Array().Type().(mir.ArrayType))
			ptr := self.builder.CreateAlloca("", at)
			self.builder.CreateStore(array, ptr)
			elemPtr := self.builder.CreateInBoundsGEP("", at, ptr, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
			return self.builder.CreateLoad("", self.codegenType(sv.Type()), elemPtr)
		}
	case *mir.StructIndex:
		st := self.codegenValue(sv.Struct())
		if sv.IsPtr(){
			stt := self.codegenStructType(sv.Struct().Type().(mir.PtrType).Elem().(mir.StructType))
			return self.builder.CreateStructGEP("", stt, st, uint(sv.Index()))
		}else{
			return self.builder.CreateExtractValue("", st, uint(sv.Index()))
		}
	case *mir.Phi:
		froms := lo.Map(sv.Froms(), func(item pair.Pair[*mir.Block, mir.Value], _ int) struct {
			Value llvm.Value
			Block llvm.Block
		} {
			return struct {
				Value llvm.Value
				Block llvm.Block
			}{
				Value: self.codegenValue(item.Second),
				Block: self.blocks.Get(item.First),
			}
		})
		return self.builder.CreatePHI("", self.codegenType(sv.Type()), froms...)
	case *mir.PackArray:
		ptr := self.builder.CreateAlloca("", self.codegenArrayType(sv.Type().(mir.ArrayType)))
		for i, elemIr := range sv.Elems() {
			elem := self.codegenValue(elemIr)
			elemPtr := self.builder.CreateInBoundsGEP("", self.codegenArrayType(sv.Type().(mir.ArrayType)), ptr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(i)))
			self.builder.CreateStore(elem, elemPtr)
		}
		return self.builder.CreateLoad("", self.codegenType(sv.Type()), ptr)
	case *mir.PackStruct:
		ptr := self.builder.CreateAlloca("", self.codegenStructType(sv.Type().(mir.StructType)))
		for i, elemIr := range sv.Elems() {
			elem := self.codegenValue(elemIr)
			elemPtr := self.builder.CreateStructGEP("", self.codegenStructType(sv.Type().(mir.StructType)), ptr, uint(i))
			self.builder.CreateStore(elem, elemPtr)
		}
		return self.builder.CreateLoad("", self.codegenType(sv.Type()), ptr)
	default:
		panic("unreachable")
	}
}
