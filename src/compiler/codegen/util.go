package codegen

import "github.com/kkkunny/llvm"

func (self *CodeGenerator) CreateCallLLVMLifetimeEnd(size, ptr llvm.Value) {
	fn := self.module.NamedFunction("llvm.lifetime.end")
	if fn.IsNil() {
		fnType := llvm.FunctionType(
			self.ctx.VoidType(),
			[]llvm.Type{self.ctx.Int64Type(), llvm.PointerType(self.ctx.Int8Type(), 0)},
			false,
		)
		fn = llvm.AddFunction(self.module, "llvm.lifetime.end", fnType)
	}
	self.builder.CreateCall(fn, []llvm.Value{size, ptr}, "")
}
