package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) CurrentFunction() llvm.Function {
	return self.CurrentBlock().Belong()
}
