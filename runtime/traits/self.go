package traits

import (
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/runtime/types"
)

var selfType = new(_selfType)

type _selfType struct {}

func (*_selfType) String() string {
	panic("unreachable")
}

func (*_selfType) Hash() uint64 {
	panic("unreachable")
}

func (*_selfType) Equal(_ types.Type) bool {
	panic("unreachable")
}

func replaceSelfType(src types.Type, to types.Type)types.Type{
	switch t := src.(type) {
	case *types.EmptyType, *types.SintType, *types.UintType, *types.FloatType, *types.BoolType, *types.StringType:
		return t
	case *types.ArrayType:
		return types.NewArrayType(t.Size, replaceSelfType(t.Elem, to))
	case *types.TupleType:
		return types.NewTupleType(stlslices.Map(t.Elems, func(_ int, e types.Type) types.Type {
			return replaceSelfType(e, to)
		})...)
	case *types.StructType:
		return types.NewStructType(
			t.Pkg,
			t.Name,
			stlslices.Map(t.Fields, func(_ int, f types.Field) types.Field{
				return types.NewField(replaceSelfType(f.Type, to), f.Name)
			}),
			t.Methods,
		)
	case *types.UnionType:
		return types.NewUnionType(stlslices.Map(t.Elems, func(_ int, e types.Type) types.Type {
			return replaceSelfType(e, to)
		})...)
	case *types.PtrType:
		return types.NewPtrType(replaceSelfType(t.Elem, to))
	case *types.FuncType:
		return types.NewFuncType(replaceSelfType(t.Ret, to), stlslices.Map(t.Params, func(_ int, e types.Type) types.Type {
			return replaceSelfType(e, to)
		})...)
	case *types.RefType:
		return types.NewPtrType(replaceSelfType(t.Elem, to))
	case *_selfType:
		return to
	default:
		panic("unreachable")
	}
}
