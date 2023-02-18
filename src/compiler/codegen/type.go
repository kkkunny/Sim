package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
)

// 类型
func (self *CodeGenerator) codegenType(mean hir.Type) llvm.Type {
	switch mean.Kind {
	case hir.TNone:
		return self.ctx.VoidType()
	case hir.TBool:
		return t_bool
	case hir.TI8, hir.TU8:
		return self.ctx.Int8Type()
	case hir.TI16, hir.TU16:
		return self.ctx.Int16Type()
	case hir.TI32, hir.TU32:
		return self.ctx.Int32Type()
	case hir.TI64, hir.TU64:
		return self.ctx.Int64Type()
	case hir.TIsize, hir.TUsize:
		return t_size
	case hir.TF32:
		return self.ctx.FloatType()
	case hir.TF64:
		return self.ctx.DoubleType()
	case hir.TPtr:
		return llvm.PointerType(self.codegenType(mean.GetPtr()), 0)
	case hir.TFunc:
		ret := self.codegenType(mean.GetFuncRet())
		paramHirs := mean.GetFuncParams()
		params := make([]llvm.Type, len(paramHirs))
		for i, p := range paramHirs {
			params[i] = self.codegenType(p)
		}
		return llvm.PointerType(llvm.FunctionType(ret, params, mean.GetFuncVarArg()), 0)
	case hir.TArray:
		elem := self.codegenType(mean.GetArrayElem())
		return llvm.ArrayType(elem, int(mean.GetArraySize()))
	case hir.TTuple:
		elemHirs := mean.GetTupleElems()
		if len(elemHirs) <= 3 {
			elems := make([]llvm.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.codegenType(e)
			}
			return self.ctx.StructType(elems, false)
		} else {
			key := mean.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			td := self.ctx.StructCreateNamed("")
			elems := make([]llvm.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.codegenType(e)
			}
			td.StructSetBody(elems, false)
			self.types[key] = td
			return td
		}
	case hir.TStruct:
		fieldHirs := mean.GetStructFields()
		if len(fieldHirs) <= 3 {
			elems := make([]llvm.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.codegenType(f.Third)
			}
			return self.ctx.StructType(elems, false)
		} else {
			key := mean.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			td := self.ctx.StructCreateNamed("")
			elems := make([]llvm.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.codegenType(f.Third)
			}
			td.StructSetBody(elems, false)
			self.types[key] = td
			return td
		}
	case hir.TTypedef:
		def := mean.GetTypedef()
		if def.Target.Kind != hir.TTuple && def.Target.Kind != hir.TStruct {
			return self.codegenType(def.Target)
		}
		key := mean.String()
		if t, ok := self.types[key]; ok {
			return t
		}
		td := self.ctx.StructCreateNamed("")
		self.types[key] = td
		switch def.Target.Kind {
		case hir.TTuple:
			elemHirs := def.Target.GetTupleElems()
			elems := make([]llvm.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.codegenType(e)
			}
			td.StructSetBody(elems, false)
		case hir.TStruct:
			fieldHirs := def.Target.GetStructFields()
			elems := make([]llvm.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.codegenType(f.Third)
			}
			td.StructSetBody(elems, false)
		default:
			panic("unreachable")
		}
		return td
	default:
		panic("unreachable")
	}
}
