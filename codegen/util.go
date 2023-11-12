package codegen

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
)

func (self *CodeGenerator) buildArrayIndex(t llvm.ArrayType, from, index llvm.Value, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		elemPtr := self.builder.CreateInBoundsGEP("", t, from, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
		if !load {
			return elemPtr
		}
		return self.builder.CreateLoad("", t.Element(), elemPtr)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractElement("", from, index)
	}
}

func (self *CodeGenerator) buildArrayIndexWith(t llvm.ArrayType, from llvm.Value, index uint, load bool) llvm.Value {
	return self.buildArrayIndex(t, from, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(index)), load)
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

func (self *CodeGenerator) buildEqual(l, r llvm.Value) llvm.Value {
	t := l.Type()

	switch tt := t.(type) {
	case llvm.IntegerType, llvm.PointerType:
		return self.builder.CreateIntCmp("", llvm.IntEQ, l, r)
	case llvm.FloatType:
		return self.builder.CreateFloatCmp("", llvm.FloatOEQ, l, r)
	case llvm.ArrayType:
		// TODO
		panic("unreachable")
	case llvm.StructType:
		st := l.Type().(llvm.StructType)
		if st.CountElems() == 0 {
			return self.ctx.ConstInteger(self.ctx.IntegerType(1), 1)
		}

		beginBlock := self.builder.CurrentBlock()
		nextBlocks := dynarray.NewDynArrayWithLength[pair.Pair[llvm.Value, llvm.Block]](uint(tt.CountElems()))
		srcBlocks := dynarray.NewDynArrayWithLength[llvm.Block](uint(tt.CountElems()))
		for i := uint32(0); i < tt.CountElems(); i++ {
			cond := self.buildEqual(self.buildStructIndex(st, l, uint(i), true), self.buildStructIndex(st, r, uint(i), true))
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
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildNotEqual(l, r llvm.Value) llvm.Value {
	t := l.Type()

	switch t.(type) {
	case llvm.IntegerType, llvm.PointerType:
		return self.builder.CreateIntCmp("", llvm.IntNE, l, r)
	case llvm.FloatType:
		return self.builder.CreateFloatCmp("", llvm.FloatUNE, l, r)
	case llvm.ArrayType, llvm.StructType:
		return self.builder.CreateNot("", self.buildEqual(l, r))
	default:
		panic("unreachable")
	}
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
