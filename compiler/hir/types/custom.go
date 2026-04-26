package types

import (
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

func NewCustomType(name string, underlying Type) CustomType {
	switch underlying := underlying.(type) {
	case SintType:
		return &_CustomSintType{
			_CustomBaseType[SintType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case UintType:
		return &_CustomUintType{
			_CustomBaseType[UintType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case FloatType:
		return &_CustomFloatType{
			_CustomBaseType[FloatType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case BooleanType:
		return &_CustomBooleanType{
			_CustomBaseType[BooleanType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case RefType:
		return &_CustomRefType{
			_CustomBaseType[RefType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case FuncType:
		return &_CustomFuncType{
			_CustomBaseType[FuncType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case ArrayType:
		return &_CustomArrayType{
			_CustomBaseType[ArrayType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case TupleType:
		return &_CustomTupleType{
			_CustomBaseType[TupleType]{
				Name:       name,
				Underlying: underlying,
			},
		}
	case UnionType:
		return &_CustomUnionType{
			_CustomBaseType[UnionType]{
				Name:       name,
				Underlying: underlying,
			},
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
