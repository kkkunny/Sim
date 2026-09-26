package types

import (
	"reflect"

	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
)

type CustomType interface {
	hir.Type
	GetName() string
	GetUnderlying() hir.Type
	GetDef() *globals.TypeDef
}

type _CustomBaseType[T hir.Type] struct {
	def *globals.TypeDef
}

func NewCustomType[T hir.Type](def *globals.TypeDef) CustomType {
	t := reflect.TypeFor[T]()
	switch {
	case t.AssignableTo(reflect.TypeFor[SintType]()):
		return &_CustomSintType{
			_CustomBaseType[SintType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[UintType]()):
		return &_CustomUintType{
			_CustomBaseType[UintType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[FloatType]()):
		return &_CustomFloatType{
			_CustomBaseType[FloatType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[BooleanType]()):
		return &_CustomBooleanType{
			_CustomBaseType[BooleanType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[StringType]()):
		return &_CustomStringType{
			_CustomBaseType[StringType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[RefType]()):
		return &_CustomRefType{
			_CustomBaseType[RefType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[FuncType]()):
		return &_CustomFuncType{
			_CustomBaseType[FuncType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[ArrayType]()):
		return &_CustomArrayType{
			_CustomBaseType[ArrayType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[TupleType]()):
		return &_CustomTupleType{
			_CustomBaseType[TupleType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[UnionType]()):
		return &_CustomUnionType{
			_CustomBaseType[UnionType]{def: def},
		}
	case t.AssignableTo(reflect.TypeFor[StructType]()):
		return &_CustomStructType{
			_CustomBaseType[StructType]{def: def},
		}
	default:
		panic("unreachable")
	}
}

func (t *_CustomBaseType[T]) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_CustomBaseType[T]) String() string {
	return t.def.Name
}

func (t *_CustomBaseType[T]) Equal(p hir.Type) bool {
	if dst, ok := p.(CustomType); ok {
		// 自定义类型按定义身份判等（GetDef 指针唯一）：同名跨包类型、不同别名
		// 不会互相混淆。此前按名字比较会把两个包里的同名类型视为同一类型。
		return t.GetDef() == dst.GetDef()
	}
	// 与内建/字面量类型比较时递归到底层，保证与「内建.Equal(自定义)」方向对称：
	// 内建类型用结构化接口匹配自定义包装（stlval.Is[...]），若这里只比名字，
	// 会出现 x.Equal(b) 与 b.Equal(x) 结果相反的矛盾（`mb as bool` 类型问题的根因）。
	return t.GetUnderlying().Equal(p)
}

func (t *_CustomBaseType[T]) GetName() string {
	return t.def.Name
}

func (t *_CustomBaseType[T]) GetUnderlying() hir.Type {
	return t.def.Underlying
}

func (t *_CustomBaseType[T]) GetDef() *globals.TypeDef {
	return t.def
}

// GetUnderlying 获取底层类型
func GetUnderlying(t hir.Type) hir.Type {
	for {
		switch tt := t.(type) {
		case CustomType:
			t = tt.GetUnderlying()
		default:
			return tt
		}
	}
}

// CheckRecursion 检查循环
func CheckRecursion(ct CustomType) bool {
	var checkFn func(stack set.Set[CustomType], t hir.Type) bool
	checkFn = func(stack set.Set[CustomType], t hir.Type) bool {
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
			return stlslices.Any(t.GetElems(), func(_ int, e hir.Type) bool {
				return checkFn(stack, e)
			})
		case UnionType:
			return stlslices.Any(t.GetElems(), func(_ int, e hir.Type) bool {
				return checkFn(stack, e)
			})
		case StructType:
			return stlslices.Any(t.GetFields(), func(_ int, f *StructField) bool {
				return checkFn(stack, f.Type)
			})
		default:
			return false
		}
	}
	return checkFn(set.StdLinkedHashSetWith[CustomType](), ct)
}
