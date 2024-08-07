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

func (self *CodeGenerator) getExternFunction(name string, t llvm.FunctionType) llvm.Function {
	fn, ok := self.module.GetFunction(name)
	if !ok {
		fn = self.module.NewFunction(name, t)
		fn.SetLinkage(llvm.ExternalLinkage)
	}
	return fn
}

func (self *CodeGenerator) buildEqual(t hir.Type, l, r llvm.Value, not bool) llvm.Value {
	switch irType := hir.ToRuntimeType(t).(type) {
	case *hir.SintType, *hir.UintType, *hir.RefType:
		return self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
	case *hir.FloatType:
		return self.builder.CreateFloatCmp("", stlbasic.Ternary(!not, llvm.FloatOEQ, llvm.FloatUNE), l, r)
	case *hir.ArrayType:
		t := self.codegenType(irType).(llvm.ArrayType)
		if t.Capacity() == 0 {
			return self.ctx.ConstBoolean(true)
		}

		indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
		self.builder.CreateStore(self.ctx.ConstZero(self.ctx.IntPtrType(self.target)), indexPtr)
		condBlock := self.builder.CurrentBlock().Belong().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentBlock().Belong().NewBlock(""), self.builder.CurrentBlock().Belong().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		cond = self.buildEqual(irType.Elem, self.buildArrayIndex(t, l, index, false), self.buildArrayIndex(t, r, index, false), false)
		bodyEndBlock := self.builder.CurrentBlock()
		actionBlock := bodyEndBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, actionBlock, outBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// out
		self.builder.MoveToAfter(outBlock)
		phi := self.builder.CreatePHI(
			"",
			self.ctx.BooleanType(),
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.ctx.ConstBoolean(true), Block: condBlock},
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.ctx.ConstBoolean(false), Block: bodyEndBlock},
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
			return self.ctx.ConstBoolean(true)
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

		phi := self.builder.CreatePHI("", self.ctx.BooleanType())
		for i, b := range srcBlocks {
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: stlbasic.Ternary[llvm.Value](i != len(srcBlocks)-1, self.ctx.ConstBoolean(false), stlslices.Last(nextBlocks).First), Block: b})
		}
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case *hir.CustomType:
		switch {
		case irType.EqualTo(self.hir.BuildinTypes.Str):
			f := self.getExternFunction("sim_runtime_str_eq_str", self.ctx.FunctionType(false, self.boolType(), l.Type(), r.Type()))
			var res llvm.Value = self.builder.CreateCall("", f.FunctionType(), f, l, r)
			if not {
				res = self.builder.CreateNot("", res)
			}
			return res
		default:
			return self.buildEqual(irType.Target, l, r, not)
		}
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
		// TODO: 枚举类型比较
		return self.ctx.ConstBoolean(true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayIndex(at llvm.ArrayType, v, i llvm.Value, expectPtr ...bool) llvm.Value {
	switch {
	case stlbasic.Is[llvm.ArrayType](v.Type()) && stlbasic.Is[llvm.ConstInteger](i):
		var value llvm.Value = self.builder.CreateInBoundsGEP("", at, self.ctx.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
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
		var value llvm.Value = self.builder.CreateInBoundsGEP("", at, v, self.ctx.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
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

func (self *CodeGenerator) getMainFunction() llvm.Function {
	mainFn, ok := self.module.GetFunction("main")
	if !ok {
		mainFn = self.module.NewFunction("main", self.ctx.FunctionType(false, self.ctx.IntegerType(8)))
		mainFn.NewBlock("")
	}
	return mainFn
}

func (self *CodeGenerator) getInitFunction() llvm.Function {
	initFn, ok := self.module.GetFunction("sim_runtime_init")
	if !ok {
		initFn = self.module.NewFunction("sim_runtime_init", self.ctx.FunctionType(false, self.ctx.VoidType()))
		self.module.AddConstructor(65535, initFn)
		initFn.NewBlock("")
	}
	return initFn
}

func (self *CodeGenerator) constStringPtr(s string) llvm.Constant {
	if !self.strings.ContainKey(s) {
		st := self.codegenType(self.hir.BuildinTypes.Str).(llvm.StructType)
		dataPtr := stlbasic.TernaryAction(s == "", func() llvm.Constant {
			return self.ctx.ConstZero(st.Elems()[0])
		}, func() llvm.Constant {
			data := self.module.NewConstant("", self.ctx.ConstString(s))
			data.SetLinkage(llvm.PrivateLinkage)
			return self.ctx.ConstInBoundsGEP(
				data.Type(),
				data,
				self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0),
				self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0),
			)
		})
		str := self.module.NewConstant("", self.ctx.ConstStruct(false, dataPtr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(len(s)))))
		str.SetLinkage(llvm.PrivateLinkage)
		self.strings.Set(s, str)
	}
	return self.strings.Get(s)
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
		elemPtr := self.builder.CreateInBoundsGEP("", at, ptr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(i)))
		self.builder.CreateStore(elem, elemPtr)
	}
	return self.builder.CreateLoad("", at, ptr)
}

func (self *CodeGenerator) buildPanic(s string) {
	fn := self.getExternFunction("sim_runtime_panic", self.ctx.FunctionType(false, self.ctx.VoidType(), self.ctx.OpaquePointerType()))
	self.builder.CreateCall("", fn.FunctionType(), fn, self.constStringPtr(s))
	self.builder.CreateUnreachable()
}

func (self *CodeGenerator) buildCheckZero(v llvm.Value) {
	var cond llvm.Value
	if stlbasic.Is[llvm.FloatType](v.Type()) {
		cond = self.builder.CreateFloatCmp("", llvm.FloatOEQ, v, self.ctx.ConstZero(v.Type()))
	} else {
		cond = self.builder.CreateIntCmp("", llvm.IntEQ, v, self.ctx.ConstZero(v.Type()))
	}
	f := self.builder.CurrentBlock().Belong()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("zero exception")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildCheckIndex(index llvm.Value, rangev uint64) {
	cond := self.builder.CreateIntCmp("", llvm.IntUGE, index, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), int64(rangev)))
	f := self.builder.CurrentBlock().Belong()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("index out of range")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildMalloc(t llvm.Type) llvm.Value {
	fn := self.getExternFunction("sim_runtime_malloc", self.ctx.FunctionType(false, self.ctx.OpaquePointerType(), self.ctx.IntPtrType(self.target)))
	size := stlmath.RoundTo(self.target.GetStoreSizeOfType(t), self.target.GetABIAlignOfType(t))
	return self.builder.CreateCall("", fn.FunctionType(), fn, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(size)))
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
		if self.builder.CurrentBlock().Belong() == method {
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
		indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
		self.builder.CreateStore(self.ctx.ConstZero(self.ctx.IntPtrType(self.target)), indexPtr)
		condBlock := self.builder.CurrentBlock().Belong().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentBlock().Belong().NewBlock(""), self.builder.CurrentBlock().Belong().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		self.buildStore(tir.Elem, self.buildArrayIndex(t, v, index, false), self.buildArrayIndex(t, newArrayPtr, index, true))
		actionBlock := self.builder.CurrentBlock().Belong().NewBlock("")
		self.builder.CreateBr(actionBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
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
			self.buildStore(fields[i], self.buildStructIndex(t, v, uint(i), false), self.buildStructIndex(t, newStructPtr, uint(i), true))
		}
		return self.builder.CreateLoad("", t, newStructPtr)
	case *hir.EnumType:
		// TODO: 枚举类型拷贝
		return v
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildStore(t hir.Type, from, to llvm.Value) llvm.Store {
	return self.builder.CreateStore(self.buildCopy(t, from), to)
}

func (self *CodeGenerator) buildReturn(t hir.Type, v ...llvm.Value) llvm.Return {
	if stlslices.Empty(v) {
		return self.builder.CreateRet(nil)
	}
	return self.builder.CreateRet(stlbasic.Ptr(self.buildCopy(t, stlslices.First(v))))
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
