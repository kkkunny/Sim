package types

import (
	"reflect"

	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

type CustomType interface {
	Type
	GetName() string
	GetUnderlying() Type
}

type _CustomBaseType[T Type] struct {
	Name       string
	Underlying T
}

func DelayNewCustomType[T Type](name string) (CustomType, func(underlying T)) {
	t := reflect.TypeFor[T]()
	switch {
	case t.AssignableTo(reflect.TypeFor[SintType]()):
		ct := &_CustomSintType{
			_CustomBaseType[SintType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, SintType](t)
		}
	case t.AssignableTo(reflect.TypeFor[UintType]()):
		ct := &_CustomUintType{
			_CustomBaseType[UintType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, UintType](t)
		}
	case t.AssignableTo(reflect.TypeFor[FloatType]()):
		ct := &_CustomFloatType{
			_CustomBaseType[FloatType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, FloatType](t)
		}
	case t.AssignableTo(reflect.TypeFor[BooleanType]()):
		ct := &_CustomBooleanType{
			_CustomBaseType[BooleanType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, BooleanType](t)
		}
	case t.AssignableTo(reflect.TypeFor[StringType]()):
		ct := &_CustomStringType{
			_CustomBaseType[StringType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, StringType](t)
		}
	case t.AssignableTo(reflect.TypeFor[RefType]()):
		ct := &_CustomRefType{
			_CustomBaseType[RefType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, RefType](t)
		}
	case t.AssignableTo(reflect.TypeFor[FuncType]()):
		ct := &_CustomFuncType{
			_CustomBaseType[FuncType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, FuncType](t)
		}
	case t.AssignableTo(reflect.TypeFor[ArrayType]()):
		ct := &_CustomArrayType{
			_CustomBaseType[ArrayType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, ArrayType](t)
		}
	case t.AssignableTo(reflect.TypeFor[TupleType]()):
		ct := &_CustomTupleType{
			_CustomBaseType[TupleType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, TupleType](t)
		}
	case t.AssignableTo(reflect.TypeFor[UnionType]()):
		ct := &_CustomUnionType{
			_CustomBaseType[UnionType]{
				Name: name,
			},
		}
		return ct, func(t T) {
			ct.Underlying = stlval.As[T, UnionType](t)
		}
	default:
		panic("unreachable")
	}
}

func (t *_CustomBaseType[T]) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_CustomBaseType[T]) String() string {
	return t.Name
}

func (t *_CustomBaseType[T]) Equal(p Type) bool {
	dst, ok := p.(CustomType)
	if !ok {
		return false
	}
	return t.Name == dst.GetName()
}

func (t *_CustomBaseType[T]) GetName() string {
	return t.Name
}

func (t *_CustomBaseType[T]) GetUnderlying() Type {
	return t.Underlying
}

func GetUnderlying(t Type) Type {
	for {
		switch tt := t.(type) {
		case CustomType:
			t = tt.GetUnderlying()
		default:
			return tt
		}
	}
}

func CheckRecursion(ct CustomType) bool {
	var checkFn func(stack set.Set[CustomType], t Type) bool
	checkFn = func(stack set.Set[CustomType], t Type) bool {
		switch t := t.(type) {
		case CustomType:
			if !stack.Add(t) {
				return true
			}
			defer stack.Remove(t)
			return checkFn(stack, t.GetUnderlying())
		case ArrayType:
			return checkFn(stack, t.GetElem())
		case TupleType:
			return stlslices.Any(t.GetElems(), func(_ int, e Type) bool {
				return checkFn(stack, e)
			})
		case UnionType:
			return stlslices.Any(t.GetElems(), func(_ int, e Type) bool {
				return checkFn(stack, e)
			})
		default:
			return false
		}
	}
	return checkFn(set.StdLinkedHashSetWith[CustomType](), ct)
}
