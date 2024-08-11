package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) GetExternFunction(name string, t llvm.FunctionType) llvm.Function {
	fn, ok := self.GetFunction(name)
	if !ok {
		fn = self.NewFunction(name, t)
		fn.SetLinkage(llvm.ExternalLinkage)
	}
	return fn
}
