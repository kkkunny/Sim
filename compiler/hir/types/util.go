package types

import (
	"github.com/kkkunny/stl/container/hashmap"
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
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType, VirtualType:
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
		} else if !stlval.Is[hir.BuildInType](tt) {
			return tt, true
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
		} else if !stlval.Is[hir.BuildInType](tt) {
			return tt, true
		}
		return wrap[T](t, tt), true
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
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType, VirtualType:
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
	default:
		panic("unreachable")
	}
}

func ReplaceVirtualType(table hashmap.HashMap[VirtualType, hir.Type], typ hir.Type) hir.Type {
	if table.Empty() {
		return typ
	}

	switch t := typ.(type) {
	case VirtualType:
		return table.Get(t, t)
	case GenericCustomType:
		return t.WithGenericArgs(stlslices.Map(t.GenericArgs(), func(_ int, arg hir.Type) hir.Type {
			return ReplaceVirtualType(table, arg)
		}))
	case NoThingType, NoReturnType, NumType, BoolType, StrType, TypeDef:
		return t
	case RefType:
		return NewRefType(t.Mutable(), ReplaceVirtualType(table, t.Pointer()))
	case ArrayType:
		return NewArrayType(ReplaceVirtualType(table, t.Elem()), t.Size())
	case TupleType:
		return NewTupleType(stlslices.Map(t.Elems(), func(_ int, elem hir.Type) hir.Type {
			return ReplaceVirtualType(table, elem)
		})...)
	case FuncType:
		return NewFuncType(ReplaceVirtualType(table, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceVirtualType(table, param)
		})...)
	case LambdaType:
		return NewLambdaType(ReplaceVirtualType(table, t.Ret()), stlslices.Map(t.Params(), func(_ int, param hir.Type) hir.Type {
			return ReplaceVirtualType(table, param)
		})...)
	case StructType:
		return NewStructType(stlslices.Map(t.Fields().Values(), func(_ int, field *Field) *Field {
			return NewField(field.pub, field.mut, field.name, ReplaceVirtualType(table, field.typ))
		})...)
	case EnumType:
		return NewEnumType(stlslices.Map(t.EnumFields().Values(), func(_ int, field *EnumField) *EnumField {
			var newElem []hir.Type
			if elem, ok := field.Elem(); ok {
				newElem = append(newElem, ReplaceVirtualType(table, elem))
			}
			return NewEnumField(field.name, newElem...)
		})...)
	default:
		panic("unreachable")
	}
}
