package llvmUtil

import (
	"github.com/kkkunny/go-llvm"
)

func (self *Builder) GetExternFunction(name string, t llvm.FunctionType) llvm.Function {
	fn, ok := self.GetFunction(name)
	if !ok {
		fn = self.NewFunction(name, t)
		fn.SetLinkage(llvm.ExternalLinkage)
	}
	return fn
}

func (self *Builder) GetMainFunction() llvm.Function {
	fn, ok := self.GetFunction("main")
	if !ok {
		fn = self.NewFunction("main", self.FunctionType(false, self.IntegerType(8)))
		fn.NewBlock("")
	}
	return fn
}

func (self *Builder) GetInitFunction() llvm.Function {
	fn, ok := self.GetFunction("sim_runtime_init")
	if !ok {
		fn = self.NewFunction("sim_runtime_init", self.FunctionType(false, self.VoidType()))
		fn.SetLinkage(llvm.PrivateLinkage)
		self.AddConstructor(65535, fn)
		fn.NewBlock("")
	}
	return fn
}
