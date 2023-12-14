package llvm

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

func (self *LLVMOutputer) codegenType(ir mir.Type)llvm.Type{
	switch t := ir.(type) {
	case mir.VoidType:
		return self.codegenVoidType()
	case mir.IntType:
		return self.codegenIntType(t)
	case mir.FloatType:
		return self.codegenFloatType(t)
	case mir.PtrType:
		return self.codegenPtrType(t)
	case mir.FuncType:
		return self.codegenFuncTypePtr(t)
	case mir.ArrayType:
		return self.codegenArrayType(t)
	case mir.StructType:
		return self.codegenStructType(t)
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenVoidType()llvm.VoidType{
	return self.ctx.VoidType()
}

func (self *LLVMOutputer) codegenIntType(ir mir.IntType)llvm.IntegerType{
	return self.ctx.IntegerType(uint32(ir.Bits()))
}

func (self *LLVMOutputer) codegenFloatType(ir mir.FloatType)llvm.FloatType{
	var kind llvm.FloatTypeKind
	switch ir.Bits() {
	case 16:
		kind = llvm.FloatTypeKindHalf
	case 32:
		kind = llvm.FloatTypeKindFloat
	case 64:
		kind = llvm.FloatTypeKindDouble
	case 80:
		kind = llvm.FloatTypeKindX86FP80
	case 128:
		kind = llvm.FloatTypeKindFP128
	default:
		panic("unreachable")
	}
	return self.ctx.FloatType(kind)
}

func (self *LLVMOutputer) codegenPtrType(ir mir.PtrType)llvm.PointerType{
	return self.ctx.PointerType(self.codegenType(ir.Elem()))
}

func (self *LLVMOutputer) codegenFuncType(ir mir.FuncType)llvm.FunctionType{
	switch {
	case self.target.IsWindows():
		ret := self.codegenType(ir.Ret())
		params := lo.Map(ir.Params(), func(item mir.Type, _ int) llvm.Type {
			return self.codegenType(item)
		})
		if stlbasic.Is[llvm.StructType](ret) || stlbasic.Is[llvm.ArrayType](ret) {
			params = append([]llvm.Type{self.ctx.PointerType(ret)}, params...)
			ret = self.ctx.VoidType()
		}
		for i, p := range params {
			if stlbasic.Is[llvm.StructType](p) || stlbasic.Is[llvm.ArrayType](p) {
				params[i] = self.ctx.PointerType(p)
			}
		}
		return self.ctx.FunctionType(false, ret, params...)
	default:
		params := lo.Map(ir.Params(), func(item mir.Type, _ int) llvm.Type {
			return self.codegenType(item)
		})
		return self.ctx.FunctionType(false, self.codegenType(ir.Ret()), params...)
	}
}

func (self *LLVMOutputer) codegenFuncTypePtr(ir mir.FuncType)llvm.PointerType{
	return self.ctx.PointerType(self.codegenFuncType(ir))
}

func (self *LLVMOutputer) codegenArrayType(ir mir.ArrayType)llvm.ArrayType{
	return self.ctx.ArrayType(self.codegenType(ir.Elem()), uint32(ir.Length()))
}

func (self *LLVMOutputer) codegenStructType(ir mir.StructType)llvm.StructType{
	switch st := ir.(type) {
	case *mir.UnnamedStructType:
		elems := lo.Map(st.Elems(), func(item mir.Type, _ int) llvm.Type {
			return self.codegenType(item)
		})
		return self.ctx.StructType(false, elems...)
	case *mir.NamedStruct:
		return self.types.Get(st).(llvm.StructType)
	default:
		panic("unreachable")
	}
}
