package codegen

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) buildArrayIndex(t llvm.ArrayType, from, index llvm.Value, load bool) llvm.Value {
	elemPtr := self.builder.CreateInBoundsGEP("", t, from, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
	if !load {
		return elemPtr
	}
	return self.builder.CreateLoad("", t.Element(), elemPtr)
}

func (self *CodeGenerator) buildArrayIndexWith(t llvm.ArrayType, from llvm.Value, index uint, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		return self.buildArrayIndex(t, from, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(index)), load)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractValue("", from, index)
	}
}

func (self *CodeGenerator) buildStructIndex(t llvm.StructType, from llvm.Value, index uint, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		elemPtr := self.builder.CreateStructGEP("", t, from, index)
		if !load {
			return elemPtr
		}
		return self.builder.CreateLoad("", t.GetElem(uint32(index)), elemPtr)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractValue("", from, index)
	}
}

func (self *CodeGenerator) buildEqual(t mean.Type, l, r llvm.Value, not bool) llvm.Value {
	switch meanType := t.(type) {
	case *mean.EmptyType:
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), stlbasic.Ternary[int64](!not, 1, 0))
	case mean.IntType, *mean.BoolType, *mean.FuncType:
		return self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
	case *mean.FloatType:
		return self.builder.CreateFloatCmp("", stlbasic.Ternary(!not, llvm.FloatOEQ, llvm.FloatUNE), l, r)
	case *mean.ArrayType:
		at := self.codegenArrayType(meanType)
		lp, rp := self.builder.CreateAlloca("", at), self.builder.CreateAlloca("", at)
		self.builder.CreateStore(l, lp)
		self.builder.CreateStore(r, rp)

		res := self.buildArrayEqual(meanType, lp, rp)
		if not {
			res = self.builder.CreateNot("", res)
		}
		return res
	case *mean.TupleType, *mean.StructType:
		res := self.buildStructEqual(meanType, l, r)
		if not {
			res = self.builder.CreateNot("", res)
		}
		return res
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayEqual(meanType *mean.ArrayType, l, r llvm.Value) llvm.Value {
	t := self.codegenArrayType(meanType)
	if t.Capacity() == 0 {
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), 1)
	}

	indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
	self.builder.CreateStore(self.ctx.ConstNull(self.ctx.IntPtrType(self.target)), indexPtr)
	condBlock := self.builder.CurrentBlock().Belong().NewBlock("")
	self.builder.CreateBr(condBlock)

	// cond
	self.builder.MoveToAfter(condBlock)
	var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(t.Capacity())))
	bodyBlock, outBlock := self.builder.CurrentBlock().Belong().NewBlock(""), self.builder.CurrentBlock().Belong().NewBlock("")
	self.builder.CreateCondBr(cond, bodyBlock, outBlock)

	// body
	self.builder.MoveToAfter(bodyBlock)
	index := self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr)
	cond = self.buildEqual(meanType.Elem, self.buildArrayIndex(t, l, index, true), self.buildArrayIndex(t, r, index, true), false)
	bodyEndBlock := self.builder.CurrentBlock()
	actionBlock := bodyEndBlock.Belong().NewBlock("")
	self.builder.CreateCondBr(cond, actionBlock, outBlock)

	// action
	self.builder.MoveToAfter(actionBlock)
	self.builder.CreateStore(self.builder.CreateUAdd("", index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
	self.builder.CreateBr(condBlock)

	// out
	self.builder.MoveToAfter(outBlock)
	phi := self.builder.CreatePHI("", self.ctx.IntegerType(1))
	phi.AddIncomings(
		struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.ctx.ConstInteger(self.ctx.IntegerType(1), 1), Block: condBlock},
		struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.ctx.ConstInteger(self.ctx.IntegerType(1), 0), Block: bodyEndBlock},
	)
	return phi
}

func (self *CodeGenerator) buildStructEqual(meanType mean.Type, l, r llvm.Value) llvm.Value {
	_, isTuple := meanType.(*mean.TupleType)
	t := stlbasic.TernaryAction(isTuple, func() llvm.StructType {
		return self.codegenTupleType(meanType.(*mean.TupleType))
	}, func() llvm.StructType {
		return self.codegenStructType(meanType.(*mean.StructType))
	})
	fields := stlbasic.TernaryAction(isTuple, func() dynarray.DynArray[mean.Type] {
		return dynarray.NewDynArrayWith(meanType.(*mean.TupleType).Elems...)
	}, func() dynarray.DynArray[mean.Type] {
		return meanType.(*mean.StructType).Fields.Values()
	})

	if t.CountElems() == 0 {
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), 1)
	}

	beginBlock := self.builder.CurrentBlock()
	nextBlocks := dynarray.NewDynArrayWithLength[pair.Pair[llvm.Value, llvm.Block]](uint(t.CountElems()))
	srcBlocks := dynarray.NewDynArrayWithLength[llvm.Block](uint(t.CountElems()))
	for i := uint32(0); i < t.CountElems(); i++ {
		cond := self.buildEqual(fields.Get(uint(i)), self.buildStructIndex(t, l, uint(i), true), self.buildStructIndex(t, r, uint(i), true), false)
		nextBlock := beginBlock.Belong().NewBlock("")
		nextBlocks.Set(uint(i), pair.NewPair(cond, nextBlock))
		srcBlocks.Set(uint(i), self.builder.CurrentBlock())
		self.builder.MoveToAfter(nextBlock)
	}
	self.builder.MoveToAfter(beginBlock)

	endBlock := nextBlocks.Back().Second
	for iter := nextBlocks.Iterator(); iter.Next(); {
		item := iter.Value()
		if iter.HasNext() {
			self.builder.CreateCondBr(item.First, item.Second, endBlock)
		} else {
			self.builder.CreateBr(endBlock)
		}
		self.builder.MoveToAfter(item.Second)
	}

	phi := self.builder.CreatePHI("", self.ctx.IntegerType(1))
	for iter := srcBlocks.Iterator(); iter.Next(); {
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: stlbasic.Ternary[llvm.Value](iter.HasNext(), self.ctx.ConstInteger(self.ctx.IntegerType(1), 0), nextBlocks.Back().First), Block: iter.Value()})
	}
	return phi
}

func (self *CodeGenerator) getMainFunction() llvm.Function {
	mainFnPtr := self.module.GetFunction("main")
	if mainFnPtr == nil {
		mainFn := self.module.NewFunction("main", self.ctx.FunctionType(self.ctx.IntegerType(8), nil, false))
		mainFn.NewBlock("entry")
		mainFnPtr = &mainFn
	}
	return *mainFnPtr
}
