package mir

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/samber/lo"
)

// Const 常量
type Const interface {
	Value
	IsZero()bool
	constant()
}

// Sint 有符号整数
type Sint struct {
	t SintType
	value *big.Int
}

func NewSint(t SintType, value *big.Int)*Sint{
	return &Sint{
		t: t,
		value: value,
	}
}

func (self *Sint) String()string{
	return self.value.String()
}

func (self *Sint) Type()Type{
	return self.t
}

func (self *Sint) IsZero()bool{
	return self.value.Cmp(big.NewInt(0)) == 0
}

func (*Sint)constant(){}

// Uint 无符号整数
type Uint struct {
	t UintType
	value *big.Int
}

func NewUint(t UintType, value *big.Int)*Uint{
	return &Uint{
		t: t,
		value: value,
	}
}

func (self *Uint) String()string{
	return self.value.String()
}

func (self *Uint) Type()Type{
	return self.t
}

func (self *Uint) IsZero()bool{
	return self.value.Cmp(big.NewInt(0)) == 0
}

func (*Uint)constant(){}

// Float 浮点数
type Float struct {
	t FloatType
	value *big.Float
}

func NewFloat(t FloatType, value *big.Float)*Float{
	return &Float{
		t: t,
		value: value,
	}
}

func (self *Float) String()string{
	return self.value.String()
}

func (self *Float) Type()Type{
	return self.t
}

func (self *Float) IsZero()bool{
	return self.value.Cmp(big.NewFloat(0)) == 0
}

func (*Float)constant(){}

// EmptyArray 空数组
type EmptyArray struct {
	t ArrayType
}

func NewEmptyArray(t ArrayType)*EmptyArray{
	return &EmptyArray{t: t}
}

func (self *EmptyArray) String()string{
	return "[]"
}

func (self *EmptyArray) Type()Type{
	return self.t
}

func (self *EmptyArray) IsZero()bool{
	return true
}

func (*EmptyArray)constant(){}

// EmptyStruct 空结构体
type EmptyStruct struct {
	t StructType
}

func NewEmptyStruct(t StructType)*EmptyStruct{
	return &EmptyStruct{t: t}
}

func (self *EmptyStruct) String()string{
	return "{}"
}

func (self *EmptyStruct) Type()Type{
	return self.t
}

func (self *EmptyStruct) IsZero()bool{
	return true
}

func (*EmptyStruct)constant(){}

// EmptyFunc 空函数
type EmptyFunc struct {
	t FuncType
}

func NewEmptyFunc(t FuncType)*EmptyFunc{
	return &EmptyFunc{t: t}
}

func (self *EmptyFunc) String()string{
	return "nullfunc"
}

func (self *EmptyFunc) Type()Type{
	return self.t
}

func (self *EmptyFunc) IsZero()bool{
	return true
}

func (*EmptyFunc)constant(){}

// EmptyPtr 空指针
type EmptyPtr struct {
	t PtrType
}

func NewEmptyPtr(t PtrType)*EmptyPtr{
	return &EmptyPtr{t: t}
}

func (self *EmptyPtr) String()string{
	return "nullptr"
}

func (self *EmptyPtr) Type()Type{
	return self.t
}

func (self *EmptyPtr) IsZero()bool{
	return true
}

func (*EmptyPtr)constant(){}

// NewZero 零值
func NewZero(t Type)Const{
	switch tt := t.(type) {
	case SintType:
		return NewSint(tt, big.NewInt(0))
	case UintType:
		return NewUint(tt, big.NewInt(0))
	case FloatType:
		return NewFloat(tt, big.NewFloat(0))
	case PtrType:
		return NewEmptyPtr(tt)
	case ArrayType:
		return NewEmptyArray(tt)
	case StructType:
		return NewEmptyStruct(tt)
	case FuncType:
		return NewEmptyFunc(tt)
	default:
		panic("unreachable")
	}
}

// Array 数组
type Array struct {
	t ArrayType
	elems []Const
}

func NewArray(t ArrayType, elem ...Const)Const{
	if t.Length() != uint(len(elem)){
		panic("unreachable")
	}
	for _, e := range elem{
		if !e.Type().Equal(t.Elem()){
			panic("unreachable")
		}
	}

	zero := true
	for _, e := range elem{
		if !e.IsZero(){
			zero = false
			break
		}
	}
	if zero{
		return NewEmptyArray(t)
	}

	return &Array{
		t: t,
		elems: elem,
	}
}

func (self *Array) String()string{
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("[%s]", strings.Join(elems, ","))
}

func (self *Array) Type()Type{
	return self.t
}

func (self *Array) IsZero()bool{
	return false
}

func (*Array)constant(){}

// Struct 结构体
type Struct struct {
	t StructType
	elems []Const
}

func NewStruct(t StructType, elem ...Const)Const{
	if len(t.Elems()) != len(elem){
		panic("unreachable")
	}
	for i, e := range t.Elems(){
		if !e.Equal(elem[i].Type()){
			panic("unreachable")
		}
	}

	zero := true
	for _, e := range elem{
		if !e.IsZero(){
			zero = false
			break
		}
	}
	if zero{
		return NewEmptyStruct(t)
	}

	return &Struct{
		t: t,
		elems: elem,
	}
}

func (self *Struct) String()string{
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("{%s}", strings.Join(elems, ","))
}

func (self *Struct) Type()Type{
	return self.t
}

func (self *Struct) IsZero()bool{
	return false
}

func (*Struct)constant(){}
