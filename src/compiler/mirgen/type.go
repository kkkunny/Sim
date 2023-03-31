package mirgen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/mir"
)

var (
	t_void  = mir.NewTypeVoid()
	t_bool  = mir.NewTypeSint(1)
	t_i8    = mir.NewTypeSint(1)
	t_u8    = mir.NewTypeUint(1)
	t_i16   = mir.NewTypeSint(2)
	t_u16   = mir.NewTypeUint(2)
	t_i32   = mir.NewTypeSint(4)
	t_u32   = mir.NewTypeUint(4)
	t_i64   = mir.NewTypeSint(8)
	t_u64   = mir.NewTypeUint(8)
	t_isize = mir.NewTypeSint(0)
	t_usize = mir.NewTypeUint(0)
	t_f32   = mir.NewTypeFloat(4)
	t_f64   = mir.NewTypeFloat(8)
)

// 类型
func (self *MirGenerator) genType(ir hir.Type) mir.Type {
	switch ir.Kind {
	case hir.TNone:
		return t_void
	case hir.TBool:
		return t_bool
	case hir.TI8:
		return t_i8
	case hir.TU8:
		return t_u8
	case hir.TI16:
		return t_i16
	case hir.TU16:
		return t_u16
	case hir.TI32:
		return t_i32
	case hir.TU32:
		return t_u32
	case hir.TI64:
		return t_i64
	case hir.TU64:
		return t_u64
	case hir.TIsize:
		return t_isize
	case hir.TUsize:
		return t_usize
	case hir.TF32:
		return t_f32
	case hir.TF64:
		return t_f64
	case hir.TPtr:
		return mir.NewTypePtr(self.genType(ir.GetPtr()))
	case hir.TFunc:
		paramHirs := ir.GetFuncParams()
		params := make([]mir.Type, len(paramHirs))
		for i, p := range paramHirs {
			params[i] = self.genType(p)
		}
		return mir.NewTypeFunc(ir.GetFuncVarArg(), self.genType(ir.GetFuncRet()), params...)
	case hir.TArray:
		return mir.NewTypeArray(ir.GetArraySize(), self.genType(ir.GetArrayElem()))
	case hir.TTuple:
		elemHirs := ir.GetTupleElems()
		if len(elemHirs) <= 3 {
			elems := make([]mir.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.genType(e)
			}
			return mir.NewTypeStruct(elems...)
		} else {
			key := ir.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			t := mir.NewTypeAlias(self.pkg.NewAlias("", t_void))
			self.types[key] = t
			elems := make([]mir.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.genType(e)
			}
			t.GetAlias().Target = mir.NewTypeStruct(elems...)
			return t
		}
	case hir.TStruct:
		fieldHirs := ir.GetStructFields()
		if len(fieldHirs) <= 3 {
			elems := make([]mir.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.genType(f.Third)
			}
			return mir.NewTypeStruct(elems...)
		} else {
			key := ir.String()
			if t, ok := self.types[key]; ok {
				return t
			}
			t := mir.NewTypeAlias(self.pkg.NewAlias("", t_void))
			self.types[key] = t
			elems := make([]mir.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.genType(f.Third)
			}
			t.GetAlias().Target = mir.NewTypeStruct(elems...)
			return t
		}
	case hir.TTypedef:
		def := ir.GetTypedef()
		if def.Target.Kind != hir.TTuple && def.Target.Kind != hir.TStruct {
			return self.genType(def.Target)
		}
		key := ir.String()
		if t, ok := self.types[key]; ok {
			return t
		}
		t := mir.NewTypeAlias(self.pkg.NewAlias("", t_bool))
		self.types[key] = t
		switch def.Target.Kind {
		case hir.TTuple:
			elemHirs := def.Target.GetTupleElems()
			elems := make([]mir.Type, len(elemHirs))
			for i, e := range elemHirs {
				elems[i] = self.genType(e)
			}
			t.GetAlias().Target = mir.NewTypeStruct(elems...)
		case hir.TStruct:
			fieldHirs := def.Target.GetStructFields()
			elems := make([]mir.Type, len(fieldHirs))
			for i, f := range fieldHirs {
				elems[i] = self.genType(f.Third)
			}
			t.GetAlias().Target = mir.NewTypeStruct(elems...)
		default:
			panic("unreachable")
		}
		return t
	case hir.TUnion:
		elemHirs := ir.GetUnionElems()
		elems := make([]mir.Type, len(elemHirs))
		for i, e := range elemHirs {
			elems[i] = self.genType(e)
		}
		return mir.NewTypeUnion(elems...)
	default:
		panic("unreachable")
	}
}
