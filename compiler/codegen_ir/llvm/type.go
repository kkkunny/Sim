package llvmUtil

import "github.com/kkkunny/go-llvm"

func (self *Builder) Isize() llvm.IntegerType {
	return self.IntegerType(uint32(llvm.PointerSize) * 8)
}
func (self *Builder) I8() llvm.IntegerType   { return self.IntegerType(8) }
func (self *Builder) I16() llvm.IntegerType  { return self.IntegerType(16) }
func (self *Builder) I32() llvm.IntegerType  { return self.IntegerType(32) }
func (self *Builder) I64() llvm.IntegerType  { return self.IntegerType(64) }
func (self *Builder) I128() llvm.IntegerType { return self.IntegerType(128) }

func (self *Builder) F16() llvm.FloatType  { return self.FloatType(llvm.FloatTypeKindHalf) }
func (self *Builder) F32() llvm.FloatType  { return self.FloatType(llvm.FloatTypeKindFloat) }
func (self *Builder) F64() llvm.FloatType  { return self.FloatType(llvm.FloatTypeKindDouble) }
func (self *Builder) F128() llvm.FloatType { return self.FloatType(llvm.FloatTypeKindFP128) }

func (self *Builder) Str() llvm.StructType {
	t := self.GetTypeByName("str")
	if t != nil {
		return *t
	}
	return self.NamedStructType("str", false, self.OpaquePointerType(), self.Isize())
}
