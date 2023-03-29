package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/llvm"
)

// 类型
func (self *CodeGenerator) codegenType(ir mir.Type) llvm.Type {
	switch ir.Kind {
	case mir.TVoid:
		return self.ctx.VoidType()
	case mir.TBool:
		return self.ctx.Int1Type()
	case mir.TSint, mir.TUint:
		if ir.GetWidth() == 0 {
			return self.targetData.IntPtrType()
		}
		return self.ctx.IntType(int(ir.GetWidth()) * 8)
	case mir.TFloat:
		switch ir.GetWidth() {
		case 4:
			return self.ctx.FloatType()
		case 8:
			return self.ctx.DoubleType()
		default:
			panic("unreachable")
		}
	case mir.TPtr:
		return llvm.PointerType(self.codegenType(ir.GetPtr()), 0)
	case mir.TFunc:
		ret := self.codegenType(ir.GetFuncRet())
		paramMirs := ir.GetFuncParams()
		params := make([]llvm.Type, len(paramMirs))
		for i, p := range paramMirs {
			params[i] = self.codegenType(p)
		}
		return llvm.PointerType(llvm.FunctionType(ret, params, ir.GetFuncVarArg()), 0)
	case mir.TArray:
		return llvm.ArrayType(self.codegenType(ir.GetArrayElem()), int(ir.GetArraySize()))
	case mir.TStruct:
		elemMirs := ir.GetStructElems()
		elems := make([]llvm.Type, len(elemMirs))
		for i, e := range elemMirs {
			elems[i] = self.codegenType(e)
		}
		return self.ctx.StructType(elems, false)
	case mir.TUnion:
		return llvm.ArrayType(self.ctx.Int8Type(), int(ir.Size()))
	case mir.TAlias:
		return self.types[ir.GetAlias()]
	default:
		panic("unreachable")
	}
}
