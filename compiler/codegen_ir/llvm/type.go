package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) IntPtrType() llvm.IntegerType {
	return self.Context.IntPtrType(self.Target)
}
