package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlerror "github.com/kkkunny/stl/error"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/analyse"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (self *CodeGenerator) buildEqual(t hir.Type, l, r llvm.Value, not bool) llvm.Value {
	switch irType := hir.ToRuntimeType(t).(type) {
	case *hir.SintType, *hir.UintType, *hir.RefType:
		return self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
	case *hir.FloatType:
		return self.builder.CreateFloatCmp("", stlbasic.Ternary(!not, llvm.FloatOEQ, llvm.FloatUNE), l, r)
	case *hir.ArrayType:
		t := self.codegenType(irType).(llvm.ArrayType)
		if t.Capacity() == 0 {
			return self.builder.ConstBoolean(true)
		}

		indexPtr := self.builder.CreateAlloca("", self.builder.IntPtrType())
		self.builder.CreateStore(self.builder.ConstZero(self.builder.IntPtrType()), indexPtr)
		condBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.builder.IntPtrType(), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIntPtr(int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentFunction().NewBlock(""), self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		cond = self.buildEqual(irType.Elem, self.buildArrayIndex(t, l, index, false), self.buildArrayIndex(t, r, index, false), false)
		bodyEndBlock := self.builder.CurrentBlock()
		actionBlock := bodyEndBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, actionBlock, outBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIntPtr(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// out
		self.builder.MoveToAfter(outBlock)
		phi := self.builder.CreatePHI(
			"",
			self.builder.BooleanType(),
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstBoolean(true), Block: condBlock},
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstBoolean(false), Block: bodyEndBlock},
		)
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case *hir.TupleType, *hir.StructType:
		_, isTuple := irType.(*hir.TupleType)
		t := self.codegenType(irType).(llvm.StructType)
		fields := stlbasic.TernaryAction(isTuple, func() []hir.Type {
			return irType.(*hir.TupleType).Elems
		}, func() []hir.Type {
			values := irType.(*hir.StructType).Fields.Values()
			res := make([]hir.Type, values.Length())
			var i uint
			for iter := values.Iterator(); iter.Next(); {
				res[i] = iter.Value().Type
				i++
			}
			return res
		})

		if len(t.Elems()) == 0 {
			return self.builder.ConstBoolean(true)
		}

		beginBlock := self.builder.CurrentBlock()
		nextBlocks := make([]pair.Pair[llvm.Value, llvm.Block], uint(len(t.Elems())))
		srcBlocks := make([]llvm.Block, uint(len(t.Elems())))
		for i := uint(0); i < uint(len(t.Elems())); i++ {
			cond := self.buildEqual(fields[i], self.buildStructIndex(t, l, i, false), self.buildStructIndex(t, r, i, false), false)
			nextBlock := beginBlock.Belong().NewBlock("")
			nextBlocks[i] = pair.NewPair(cond, nextBlock)
			srcBlocks[i] = self.builder.CurrentBlock()
			self.builder.MoveToAfter(nextBlock)
		}
		self.builder.MoveToAfter(beginBlock)

		endBlock := stlslices.Last(nextBlocks).Second
		for i, p := range nextBlocks {
			if i != len(nextBlocks)-1 {
				self.builder.CreateCondBr(p.First, p.Second, endBlock)
			} else {
				self.builder.CreateBr(endBlock)
			}
			self.builder.MoveToAfter(p.Second)
		}

		phi := self.builder.CreatePHI("", self.builder.BooleanType())
		for i, b := range srcBlocks {
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: stlbasic.Ternary[llvm.Value](i != len(srcBlocks)-1, self.builder.ConstBoolean(false), stlslices.Last(nextBlocks).First), Block: b})
		}
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case *hir.CustomType:
		return self.buildEqual(irType.Target, l, r, not)
	case *hir.LambdaType:
		st := self.codegenType(irType).(llvm.StructType)
		f1p := self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.buildStructIndex(st, l, 0, false), self.buildStructIndex(st, r, 0, false))
		f2p := self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.buildStructIndex(st, l, 1, false), self.buildStructIndex(st, r, 1, false))
		pp := self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.buildStructIndex(st, l, 2, false), self.buildStructIndex(st, r, 2, false))
		return stlbasic.TernaryAction(!not, func() llvm.Value {
			return self.builder.CreateAnd("", self.builder.CreateAnd("", f1p, f2p), pp)
		}, func() llvm.Value {
			return self.builder.CreateOr("", self.builder.CreateOr("", f1p, f2p), pp)
		})
	case *hir.EnumType:
		if irType.IsSimple() {
			return self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
		}

		t := self.codegenType(irType).(llvm.StructType)
		beginBlock := self.builder.CurrentBlock()
		li, ri := self.buildStructIndex(t, l, 1, false), self.buildStructIndex(t, r, 1, false)
		indexEq := self.builder.CreateIntCmp("", llvm.IntEQ, li, ri)
		bodyBlock := beginBlock.Belong().NewBlock("")
		self.builder.MoveToAfter(bodyBlock)

		values := make([]llvm.Value, irType.Fields.Length())
		indexBlockPairs := stlslices.Map(irType.Fields.Values().ToSlice(), func(i int, f hir.EnumField) struct {
			Value llvm.Value
			Block llvm.Block
		} {
			if f.Elem.IsNone() {
				return struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: self.builder.ConstInteger(li.Type().(llvm.IntegerType), int64(i)), Block: bodyBlock}
			}

			filedBlock := beginBlock.Belong().NewBlock("")
			self.builder.MoveToAfter(filedBlock)

			ldp, rdp := self.buildStructIndex(t, l, 0, true), self.buildStructIndex(t, r, 0, true)
			ftIr := f.Elem.MustValue()
			ft := self.codegenType(ftIr)
			values[i] = self.buildEqual(ftIr, self.builder.CreateLoad("", ft, ldp), self.builder.CreateLoad("", ft, rdp), false)
			return struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstInteger(li.Type().(llvm.IntegerType), int64(i)), Block: filedBlock}
		})

		endBlock := beginBlock.Belong().NewBlock("")
		for i, p := range indexBlockPairs {
			if p.Block != bodyBlock {
				continue
			}
			p.Block = endBlock
			indexBlockPairs[i] = p
		}
		self.builder.MoveToAfter(beginBlock)
		self.builder.CreateCondBr(indexEq, bodyBlock, endBlock)

		self.builder.MoveToAfter(bodyBlock)
		self.builder.CreateSwitch(li, endBlock, indexBlockPairs...)
		self.builder.MoveToAfter(endBlock)

		phi := self.builder.CreatePHI("", self.builder.BooleanType())
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.builder.ConstBoolean(false), Block: beginBlock})
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.builder.ConstBoolean(true), Block: bodyBlock})
		for i, p := range indexBlockPairs {
			if p.Block == endBlock {
				continue
			}
			self.builder.MoveToAfter(p.Block)
			self.builder.CreateBr(endBlock)
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: values[i], Block: p.Block})
			self.builder.MoveToAfter(endBlock)
		}
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayIndex(at llvm.ArrayType, v, i llvm.Value, expectPtr ...bool) llvm.Value {
	switch {
	case stlbasic.Is[llvm.ArrayType](v.Type()) && stlbasic.Is[llvm.ConstInteger](i):
		var value llvm.Value = self.builder.CreateInBoundsGEP("", at, self.builder.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.builder.CreateLoad("", at.Elem(), value)
		}
		return value
	case stlbasic.Is[llvm.ArrayType](v.Type()):
		ptr := self.builder.CreateAlloca("", at)
		self.builder.CreateStore(v, ptr)
		v = ptr
		fallthrough
	default:
		var value llvm.Value = self.builder.CreateInBoundsGEP("", at, v, self.builder.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.builder.CreateLoad("", at.Elem(), value)
		}
		return value
	}
}

func (self *CodeGenerator) buildStructIndex(st llvm.StructType, v llvm.Value, i uint, expectPtr ...bool) llvm.Value {
	if stlbasic.Is[llvm.StructType](v.Type()) {
		var value llvm.Value = self.builder.CreateExtractValue("", v, i)
		if !stlslices.Empty(expectPtr) && stlslices.Last(expectPtr) {
			ptr := self.builder.CreateAlloca("", st.GetElem(uint32(i)))
			self.builder.CreateStore(value, ptr)
			value = ptr
		}
		return value
	} else {
		var value llvm.Value = self.builder.CreateStructGEP("", st, v, i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.builder.CreateLoad("", st.GetElem(uint32(i)), value)
		}
		return value
	}
}

func (self *CodeGenerator) constString(s string) llvm.Constant {
	if !self.strings.ContainKey(s) {
		st := self.codegenType(self.hir.BuildinTypes.Str).(llvm.StructType)
		dataPtr := stlbasic.TernaryAction(s == "", func() llvm.Constant {
			return self.builder.ConstZero(st.Elems()[0])
		}, func() llvm.Constant {
			data := self.builder.NewConstant("", self.builder.ConstString(s))
			data.SetLinkage(llvm.PrivateLinkage)
			return data
		})
		self.strings.Set(s, dataPtr)
	}
	return self.builder.ConstStruct(false, self.strings.Get(s), self.builder.ConstIntPtr(int64(len(s))))
}

func (self *CodeGenerator) buildPackStruct(st llvm.StructType, elems ...llvm.Value) llvm.Value {
	ptr := self.builder.CreateAlloca("", st)
	for i, elem := range elems {
		elemPtr := self.builder.CreateStructGEP("", st, ptr, uint(i))
		self.builder.CreateStore(elem, elemPtr)
	}
	return self.builder.CreateLoad("", st, ptr)
}

func (self *CodeGenerator) buildPackArray(at llvm.ArrayType, elems ...llvm.Value) llvm.Value {
	ptr := self.builder.CreateAlloca("", at)
	for i, elem := range elems {
		elemPtr := self.builder.CreateInBoundsGEP("", at, ptr, self.builder.ConstIntPtr(0), self.builder.ConstIntPtr(int64(i)))
		self.builder.CreateStore(elem, elemPtr)
	}
	return self.builder.CreateLoad("", at, ptr)
}

func (self *CodeGenerator) buildPanic(s string) {
	fn := self.builder.GetExternFunction("sim_runtime_panic", self.builder.FunctionType(false, self.builder.VoidType(), self.builder.OpaquePointerType()))
	self.builder.CreateCall("", fn.FunctionType(), fn, self.constString(s))
	self.builder.CreateUnreachable()
}

func (self *CodeGenerator) buildCheckZero(v llvm.Value) {
	var cond llvm.Value
	if stlbasic.Is[llvm.FloatType](v.Type()) {
		cond = self.builder.CreateFloatCmp("", llvm.FloatOEQ, v, self.builder.ConstZero(v.Type()))
	} else {
		cond = self.builder.CreateIntCmp("", llvm.IntEQ, v, self.builder.ConstZero(v.Type()))
	}
	f := self.builder.CurrentFunction()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("zero exception")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildCheckIndex(index llvm.Value, rangev uint64) {
	cond := self.builder.CreateIntCmp("", llvm.IntUGE, index, self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(rangev)))
	f := self.builder.CurrentFunction()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("index out of range")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildMalloc(t llvm.Type) llvm.Value {
	fn := self.builder.GetExternFunction("sim_runtime_malloc", self.builder.FunctionType(false, self.builder.OpaquePointerType(), self.builder.IntPtrType()))
	size := stlmath.RoundTo(self.builder.GetStoreSizeOfType(t), self.builder.GetABIAlignOfType(t))
	return self.builder.CreateCall("", fn.FunctionType(), fn, self.builder.ConstIntPtr(int64(size)))
}

func (self *CodeGenerator) usizeType() llvm.IntegerType {
	return self.codegenType(self.hir.BuildinTypes.Usize).(llvm.IntegerType)
}

func (self *CodeGenerator) boolType() llvm.IntegerType {
	return self.codegenType(self.hir.BuildinTypes.Bool).(llvm.IntegerType)
}

func (self *CodeGenerator) buildCopy(t hir.Type, v llvm.Value) llvm.Value {
	switch tir := hir.ToRuntimeType(t).(type) {
	case *hir.NoReturnType, *hir.NoThingType:
		panic("unreachable")
	case *hir.CustomType:
		if !self.hir.BuildinTypes.Copy.HasBeImpled(tir) {
			return self.buildCopy(tir.Target, v)
		}
		method := self.codegenExpr(hir.LoopFindMethod(tir, self.hir.BuildinTypes.Copy.FirstMethodName()).MustValue(), true).(llvm.Function)
		if self.builder.CurrentFunction() == method {
			return self.buildCopy(tir.Target, v)
		}
		return self.builder.CreateCall("", method.FunctionType(), method, v)
	case *hir.RefType, *hir.FuncType, *hir.LambdaType, *hir.SintType, *hir.UintType, *hir.FloatType:
		return v
	case *hir.ArrayType:
		t := self.codegenType(tir).(llvm.ArrayType)
		if t.Capacity() == 0 {
			return v
		}

		newArrayPtr := self.builder.CreateAlloca("", t)
		indexPtr := self.builder.CreateAlloca("", self.builder.IntPtrType())
		self.builder.CreateStore(self.builder.ConstZero(self.builder.IntPtrType()), indexPtr)
		condBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.builder.IntPtrType(), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIntPtr(int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentFunction().NewBlock(""), self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		self.builder.CreateStore(self.buildCopy(tir.Elem, self.buildArrayIndex(t, v, index, false)), self.buildArrayIndex(t, newArrayPtr, index, true))
		actionBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(actionBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIntPtr(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// out
		self.builder.MoveToAfter(outBlock)
		return self.builder.CreateLoad("", t, newArrayPtr)
	case *hir.TupleType, *hir.StructType:
		t := self.codegenType(t).(llvm.StructType)
		fields := stlbasic.TernaryAction(stlbasic.Is[*hir.TupleType](tir), func() []hir.Type {
			return tir.(*hir.TupleType).Elems
		}, func() []hir.Type {
			values := tir.(*hir.StructType).Fields.Values()
			res := make([]hir.Type, values.Length())
			var i uint
			for iter := values.Iterator(); iter.Next(); {
				res[i] = iter.Value().Type
				i++
			}
			return res
		})

		if len(t.Elems()) == 0 {
			return v
		}

		newStructPtr := self.builder.CreateAlloca("", t)
		for i := range t.Elems() {
			self.builder.CreateStore(self.buildCopy(fields[i], self.buildStructIndex(t, v, uint(i), false)), self.buildStructIndex(t, newStructPtr, uint(i), true))
		}
		return self.builder.CreateLoad("", t, newStructPtr)
	case *hir.EnumType:
		if tir.IsSimple() {
			return v
		}

		t := self.codegenType(tir).(llvm.StructType)
		beginBlock := self.builder.CurrentBlock()
		index := self.buildStructIndex(t, v, 1, false)

		values := make([]llvm.Value, tir.Fields.Length())
		indexBlockPairs := stlslices.Map(tir.Fields.Values().ToSlice(), func(i int, f hir.EnumField) struct {
			Value llvm.Value
			Block llvm.Block
		} {
			if f.Elem.IsNone() {
				return struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(i)), Block: beginBlock}
			}

			filedBlock := beginBlock.Belong().NewBlock("")
			self.builder.MoveToAfter(filedBlock)

			data := self.buildStructIndex(t, v, 0, false)
			ptr := self.builder.CreateAlloca("", t)
			self.builder.CreateStore(self.buildCopy(f.Elem.MustValue(), data), self.buildStructIndex(t, ptr, 0, true))
			self.builder.CreateStore(index, self.buildStructIndex(t, ptr, 1, true))
			values[i] = self.builder.CreateLoad("", t, ptr)
			return struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(i)), Block: filedBlock}
		})

		endBlock := beginBlock.Belong().NewBlock("")
		for i, p := range indexBlockPairs {
			if p.Block != beginBlock {
				continue
			}
			p.Block = endBlock
			indexBlockPairs[i] = p
		}
		self.builder.MoveToAfter(beginBlock)
		self.builder.CreateSwitch(index, endBlock, indexBlockPairs...)
		self.builder.MoveToAfter(endBlock)

		phi := self.builder.CreatePHI("", v.Type())
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: v, Block: beginBlock})
		for i, p := range indexBlockPairs {
			if p.Block == endBlock {
				continue
			}
			self.builder.MoveToAfter(p.Block)
			self.builder.CreateBr(endBlock)
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: values[i], Block: p.Block})
			self.builder.MoveToAfter(endBlock)
		}
		return phi
	default:
		panic("unreachable")
	}
}

// CodegenIr 中间代码生成
func CodegenIr(target llvm.Target, path stlos.FilePath) (llvm.Module, stlerror.Error) {
	means, err := analyse.Analyse(path)
	if err != nil {
		return llvm.Module{}, err
	}
	module := New(target, means).Codegen()
	passOption := llvm.NewPassOption()
	defer passOption.Free()
	err = stlerror.ErrorWrap(module.RunPasses(passOption, append(modulePasses, functionPasses...)...))
	return module, err
}
