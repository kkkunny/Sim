package types

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

func As[T hir.Type](typ hir.Type, strict ...bool) (T, bool) {
	if typ == nil {
		return stlval.Default[T](), false
	}
	isStrict := stlslices.Last(strict)
	switch t := typ.(type) {
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType, SelfType:
		to, ok := t.(T)
		return to, ok
	case CustomType:
		to, ok := t.(T)
		if ok {
			return to, ok
		}
		if isStrict {
			return to, ok
		}
		tt, ok := As[T](t.Target())
		if !ok {
			return tt, ok
		}
		return wrap[T](t, tt), true
	case AliasType:
		to, ok := t.(T)
		if ok {
			return to, ok
		}
		tt, ok := As[T](t.Target())
		if !ok {
			return tt, ok
		}
		return wrap[T](t, tt), true
	case CompileParamType:
		to, ok := t.(T)
		return to, ok
	default:
		panic("unreachable")
	}
}

func Is[T hir.Type](typ hir.Type, strict ...bool) bool {
	if typ == nil {
		return false
	}
	isStrict := stlslices.Last(strict)
	switch t := typ.(type) {
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType, SelfType:
		return stlval.Is[T](t)
	case CustomType:
		if stlval.Is[T](t) {
			return true
		}
		if isStrict {
			return false
		}
		return Is[T](t.Target())
	case AliasType:
		if stlval.Is[T](t) {
			return true
		}
		return Is[T](t.Target())
	case CompileParamType:
		return Is[T](t)
	default:
		panic("unreachable")
	}
}

func ReplaceSelfType(to, typ hir.Type) hir.Type {
	switch t := typ.(type) {
	case RefType:
		return NewRefType(t.Mutable(), ReplaceSelfType(to, t.Pointer()))
	case ArrayType:
		return NewArrayType(ReplaceSelfType(to, t.Elem()), t.Size())
	case TupleType:
		return NewTupleType(stlslices.Map(t.Elems(), func(_ int, elem hir.Type) hir.Type {
			return ReplaceSelfType(to, elem)
		})...)
	case FuncType:
		return NewFuncType(ReplaceSelfType(to, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceSelfType(to, param)
		})...)
	case LambdaType:
		return NewLambdaType(ReplaceSelfType(to, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceSelfType(to, param)
		})...)
	case StructType:
		return NewStructType(stlslices.Map(t.Fields().Values(), func(_ int, field *Field) *Field {
			return NewField(field.pub, field.mut, field.name, ReplaceSelfType(to, field.typ))
		})...)
	case EnumType:
		return NewEnumType(stlslices.Map(t.EnumFields().Values(), func(_ int, field *EnumField) *EnumField {
			var newElem []hir.Type
			if elem, ok := field.Elem(); ok {
				newElem = append(newElem, ReplaceSelfType(to, elem))
			}
			return NewEnumField(field.name, newElem...)
		})...)
	case SelfType:
		return to
	case NoThingType, NoReturnType, NumType, BoolType, StrType, TypeDef, CompileParamType:
		return t
	default:
		panic("unreachable")
	}
}

func ReplaceCompileParam(table map[CompileParamType]hir.Type, typ hir.Type) hir.Type {
	switch t := typ.(type) {
	case RefType:
		return NewRefType(t.Mutable(), ReplaceCompileParam(table, t.Pointer()))
	case ArrayType:
		return NewArrayType(ReplaceCompileParam(table, t.Elem()), t.Size())
	case TupleType:
		return NewTupleType(stlslices.Map(t.Elems(), func(_ int, elem hir.Type) hir.Type {
			return ReplaceCompileParam(table, elem)
		})...)
	case FuncType:
		return NewFuncType(ReplaceCompileParam(table, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceCompileParam(table, param)
		})...)
	case LambdaType:
		return NewLambdaType(ReplaceCompileParam(table, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceCompileParam(table, param)
		})...)
	case StructType:
		return NewStructType(stlslices.Map(t.Fields().Values(), func(_ int, field *Field) *Field {
			return NewField(field.pub, field.mut, field.name, ReplaceCompileParam(table, field.typ))
		})...)
	case EnumType:
		return NewEnumType(stlslices.Map(t.EnumFields().Values(), func(_ int, field *EnumField) *EnumField {
			var newElem []hir.Type
			if elem, ok := field.Elem(); ok {
				newElem = append(newElem, ReplaceCompileParam(table, elem))
			}
			return NewEnumField(field.name, newElem...)
		})...)
	case CompileParamType:
		to, ok := table[t]
		if !ok {
			return t
		}
		return to
	case NoThingType, NoReturnType, NumType, BoolType, StrType, TypeDef, SelfType:
		return t
	default:
		panic("unreachable")
	}
}
