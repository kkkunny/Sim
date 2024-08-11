package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) ConstIntPtr(v int64) llvm.ConstInteger {
	return self.Context.ConstInteger(self.IntPtrType(), v)
}
