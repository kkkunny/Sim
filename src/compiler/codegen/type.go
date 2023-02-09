package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
)

// 类型
func (self *CodeGenerator) codegenType(mean hir.Type) llvm.Type {
	switch typ := mean.(type) {
	case *hir.TypeFunc:
		ret := self.codegenType(typ.Ret)
		params := make([]llvm.Type, len(typ.Params))
		for i, p := range typ.Params {
			params[i] = self.codegenType(p)
		}
		return llvm.PointerType(llvm.FunctionType(ret, params, typ.VarArg), 0)
	case *hir.TypeArray:
		elem := self.codegenType(typ.Elem)
		return llvm.ArrayType(elem, int(typ.Size))
	case *hir.TypeTuple:
		if len(typ.Elems) <= 3 {
			elems := make([]llvm.Type, len(typ.Elems))
			for i, e := range typ.Elems {
				elems[i] = self.codegenType(e)
			}
			return self.ctx.StructType(elems, false)
		} else {
			key := typ.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			td := self.ctx.StructCreateNamed("")
			elems := make([]llvm.Type, len(typ.Elems))
			for i, e := range typ.Elems {
				elems[i] = self.codegenType(e)
			}
			td.StructSetBody(elems, false)
			self.types[key] = td
			return td
		}
	case *hir.TypeStruct:
		if typ.Fields.Length() <= 3 {
			elems := make([]llvm.Type, typ.Fields.Length())
			for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
				elems[iter.Index()] = self.codegenType(iter.Value().Second)
			}
			return self.ctx.StructType(elems, false)
		} else {
			key := typ.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			td := self.ctx.StructCreateNamed("")
			elems := make([]llvm.Type, typ.Fields.Length())
			for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
				elems[iter.Index()] = self.codegenType(iter.Value().Second)
			}
			td.StructSetBody(elems, false)
			self.types[key] = td
			return td
		}
	case *hir.TypePtr:
		return llvm.PointerType(self.codegenType(typ.Elem), 0)
	case *hir.Typedef:
		if !hir.IsTupleType(typ.Dst) && !hir.IsStructType(typ.Dst) {
			return self.codegenType(typ.Dst)
		}
		key := typ.String()
		if t, ok := self.types[key]; ok {
			return t
		}
		td := self.ctx.StructCreateNamed("")
		self.types[key] = td
		switch dst := typ.Dst.(type) {
		case *hir.TypeTuple:
			elems := make([]llvm.Type, len(dst.Elems))
			for i, e := range dst.Elems {
				elems[i] = self.codegenType(e)
			}
			td.StructSetBody(elems, false)
		case *hir.TypeStruct:
			elems := make([]llvm.Type, dst.Fields.Length())
			for iter := dst.Fields.Begin(); iter.HasValue(); iter.Next() {
				elems[iter.Index()] = self.codegenType(iter.Value().Second)
			}
			td.StructSetBody(elems, false)
		default:
			panic("")
		}
		return td
	case *hir.TypeInterface:
		key := typ.String()
		if t, ok := self.types[key]; ok {
			return t
		}
		td := self.ctx.StructCreateNamed("")
		elems := make([]llvm.Type, typ.Fields.Length()+2)
		elems[0] = llvm.PointerType(self.ctx.Int8Type(), 0) // 类型
		elems[1] = llvm.PointerType(self.ctx.Int8Type(), 0) // self
		i := 2
		for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
			ft := self.codegenType(iter.Value()).ElementType()
			elems[i] = llvm.PointerType(
				llvm.FunctionType(
					ft.ReturnType(),
					append([]llvm.Type{elems[1]}, ft.ParamTypes()...),
					ft.IsFunctionVarArg(),
				), 0,
			)
			i++
		}
		td.StructSetBody(elems, false)
		self.types[key] = td
		return td
	default:
		switch {
		case hir.IsNoneType(typ):
			return self.ctx.VoidType()
		case hir.IsIntType(typ):
			switch typ {
			case hir.I8, hir.U8:
				return self.ctx.Int8Type()
			case hir.I16, hir.U16:
				return self.ctx.Int16Type()
			case hir.I32, hir.U32:
				return self.ctx.Int32Type()
			case hir.I64, hir.U64:
				return self.ctx.Int64Type()
			case hir.Isize, hir.Usize:
				return t_size
			default:
				panic("")
			}
		case hir.IsFloatType(typ):
			switch typ {
			case hir.F32:
				return self.ctx.FloatType()
			case hir.F64:
				return self.ctx.DoubleType()
			default:
				panic("")
			}
		case hir.IsBoolType(typ):
			return t_bool
		default:
			panic("")
		}
	}
}
