package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/go-llvm"
)

// 类型
func (self *CodeGenerator) codegenType(mean analyse.Type) llvm.Type {
	switch typ := mean.(type) {
	case *analyse.TypeFunc:
		ret := self.codegenType(typ.Ret)
		params := make([]llvm.Type, len(typ.Params))
		for i, p := range typ.Params {
			params[i] = self.codegenType(p)
		}
		return llvm.PointerType(llvm.FunctionType(ret, params, false), 0)
	case *analyse.TypeArray:
		elem := self.codegenType(typ.Elem)
		return llvm.ArrayType(elem, int(typ.Size))
	case *analyse.TypeTuple:
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
	case *analyse.TypeStruct:
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
	case *analyse.TypePtr:
		return llvm.PointerType(self.codegenType(typ.Elem), 0)
	case *analyse.Typedef:
		if !analyse.IsTupleType(typ.Dst) && !analyse.IsStructType(typ.Dst) {
			return self.codegenType(typ.Dst)
		}
		key := typ.String()
		if t, ok := self.types[key]; ok {
			return t
		}
		td := self.ctx.StructCreateNamed("")
		self.types[key] = td
		switch dst := typ.Dst.(type) {
		case *analyse.TypeTuple:
			elems := make([]llvm.Type, len(dst.Elems))
			for i, e := range dst.Elems {
				elems[i] = self.codegenType(e)
			}
			td.StructSetBody(elems, false)
		case *analyse.TypeStruct:
			elems := make([]llvm.Type, dst.Fields.Length())
			for iter := dst.Fields.Begin(); iter.HasValue(); iter.Next() {
				elems[iter.Index()] = self.codegenType(iter.Value().Second)
			}
			td.StructSetBody(elems, false)
		default:
			panic("")
		}
		return td
	case *analyse.TypeInterface:
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
			elems[i] = llvm.PointerType(llvm.FunctionType(ft.ReturnType(), append([]llvm.Type{elems[1]}, ft.ParamTypes()...), ft.IsFunctionVarArg()), 0)
			i++
		}
		td.StructSetBody(elems, false)
		self.types[key] = td
		return td
	default:
		switch {
		case analyse.IsNoneType(typ):
			return self.ctx.VoidType()
		case analyse.IsIntType(typ):
			switch typ {
			case analyse.I8, analyse.U8:
				return self.ctx.Int8Type()
			case analyse.I16, analyse.U16:
				return self.ctx.Int16Type()
			case analyse.I32, analyse.U32:
				return self.ctx.Int32Type()
			case analyse.I64, analyse.U64:
				return self.ctx.Int64Type()
			case analyse.Isize, analyse.Usize:
				return t_size
			default:
				panic("")
			}
		case analyse.IsFloatType(typ):
			switch typ {
			case analyse.F32:
				return self.ctx.FloatType()
			case analyse.F64:
				return self.ctx.DoubleType()
			default:
				panic("")
			}
		case analyse.IsBoolType(typ):
			return t_bool
		default:
			panic("")
		}
	}
}
