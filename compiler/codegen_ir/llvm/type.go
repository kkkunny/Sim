package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) IntPtrType() llvm.IntegerType {
	return self.Context.IntegerType(uint32(llvm.PointerSize) * 8)
}
