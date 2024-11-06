package types

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
)

func As[T Type](typ Type, strict ...bool) (T, bool) {
	if typ == nil {
		return stlval.Default[T](), false
	}
	isStrict := stlslices.Last(strict)
	switch t := typ.(type) {
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType:
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
	default:
		panic("unreachable")
	}
}

func Is[T Type](typ Type, strict ...bool) bool {
	if typ == nil {
		return false
	}
	isStrict := stlslices.Last(strict)
	switch t := typ.(type) {
	case NoThingType, NoReturnType, NumType, BoolType, StrType, RefType, ArrayType, TupleType, CallableType, StructType, EnumType:
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
