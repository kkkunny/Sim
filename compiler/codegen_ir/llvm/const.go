package llvmUtil

import (
	"github.com/kkkunny/go-llvm"
	stlval "github.com/kkkunny/stl/value"
)

func (self *Builder) ConstIsize(v int64) llvm.ConstInteger {
	return self.ConstInteger(self.Isize(), v)
}

func (self *Builder) ConstString(v string) llvm.Constant {
	if !self.stringMap.Contain(v) {
		dataPtr := stlval.TernaryAction(v == "", func() llvm.Constant {
			return self.ConstZero(self.Str().Elems()[0])
		}, func() llvm.Constant {
			data := self.NewConstant("", self.Context.ConstString(v))
			data.SetLinkage(llvm.PrivateLinkage)
			return data
		})
		self.stringMap.Set(v, dataPtr)
	}
	return self.ConstNamedStruct(self.Str(), self.stringMap.Get(v), self.ConstIsize(int64(len(v))))
}
