package llvmUtil

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
)

func (self *Builder) CurrentFunction() llvm.Function {
	return self.CurrentBlock().Belong()
}

func (self *Builder) CreateStructIndex(st llvm.StructType, v llvm.Value, i uint, expectPtr ...bool) llvm.Value {
	if stlval.Is[llvm.StructType](v.Type()) {
		var value llvm.Value = self.CreateExtractValue("", v, i)
		if !stlslices.Empty(expectPtr) && stlslices.Last(expectPtr) {
			ptr := self.CreateAlloca("", st.GetElem(uint32(i)))
			self.CreateStore(value, ptr)
			value = ptr
		}
		return value
	} else {
		var value llvm.Value = self.CreateStructGEP("", st, v, i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.CreateLoad("", st.GetElem(uint32(i)), value)
		}
		return value
	}
}

func (self *Builder) CreateArrayIndex(at llvm.ArrayType, v, i llvm.Value, expectPtr ...bool) llvm.Value {
	switch {
	case stlval.Is[llvm.ArrayType](v.Type()) && stlval.Is[llvm.ConstInteger](i):
		var value llvm.Value = self.CreateInBoundsGEP("", at, self.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.CreateLoad("", at.Elem(), value)
		}
		return value
	case stlval.Is[llvm.ArrayType](v.Type()):
		ptr := self.CreateAlloca("", at)
		self.CreateStore(v, ptr)
		v = ptr
		fallthrough
	default:
		var value llvm.Value = self.CreateInBoundsGEP("", at, v, self.ConstInteger(i.Type().(llvm.IntegerType), 0), i)
		if !stlslices.Empty(expectPtr) && !stlslices.Last(expectPtr) {
			value = self.CreateLoad("", at.Elem(), value)
		}
		return value
	}
}

func (self *Builder) CreateStruct(t llvm.StructType, elems ...llvm.Value) llvm.Value {
	ptr := self.CreateAlloca("", t)
	for i, elem := range elems {
		elemPtr := self.CreateStructGEP("", t, ptr, uint(i))
		self.CreateStore(elem, elemPtr)
	}
	return self.CreateLoad("", t, ptr)
}

func (self *Builder) CreatePackArray(t llvm.ArrayType, elems ...llvm.Value) llvm.Value {
	ptr := self.CreateAlloca("", t)
	for i, elem := range elems {
		elemPtr := self.CreateInBoundsGEP("", t, ptr, self.ConstIsize(0), self.ConstIsize(int64(i)))
		self.CreateStore(elem, elemPtr)
	}
	return self.CreateLoad("", t, ptr)
}

func (b Builder) CreateSwitch(v llvm.Value, defaultBlock llvm.Block, conds ...struct {
	Value llvm.Value
	Block llvm.Block
}) llvm.Switch {
	conds = stlslices.Filter(conds, func(_ int, cond struct {
		Value llvm.Value
		Block llvm.Block
	}) bool {
		return cond.Block != defaultBlock
	})
	return b.Builder.CreateSwitch(v, defaultBlock, conds...)
}
