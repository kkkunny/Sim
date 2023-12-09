package mir

import (
	"fmt"
	"math/big"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"
)

// Const 常量
type Const interface {
	Value
	IsZero()bool
	constant()
}

type Number interface {
	Const
	FloatValue()*big.Float
}

type Int interface {
	Number
	IntValue()*big.Int
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

func (self *Sint) Name()string{
	return self.value.String()
}

func (self *Sint) Type()Type{
	return self.t
}

func (self *Sint) IsZero()bool{
	return self.value.Cmp(big.NewInt(0)) == 0
}

func (*Sint)constant(){}

func (self *Sint)FloatValue()*big.Float{
	return new(big.Float).SetInt(self.value)
}
func (self *Sint)IntValue()*big.Int{
	return self.value
}

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

func Bool(ctx *Context, v bool)*Uint{
	if v{
		return NewUint(ctx.Bool(), big.NewInt(1))
	}else{
		return NewUint(ctx.Bool(), big.NewInt(0))
	}
}

func (self *Uint) Name()string{
	return self.value.String()
}

func (self *Uint) Type()Type{
	return self.t
}

func (self *Uint) IsZero()bool{
	return self.value.Cmp(big.NewInt(0)) == 0
}

func (*Uint)constant(){}

func (self *Uint)FloatValue()*big.Float{
	return new(big.Float).SetInt(self.value)
}
func (self *Uint)IntValue()*big.Int{
	return self.value
}

func NewInt(t IntType, value *big.Int)Int{
	switch tt := t.(type) {
	case SintType:
		return NewSint(tt, value)
	case UintType:
		return NewUint(tt, value)
	default:
		panic("unreachable")
	}
}

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

func (self *Float) Name()string{
	return self.value.String()
}

func (self *Float) Type()Type{
	return self.t
}

func (self *Float) IsZero()bool{
	return self.value.Cmp(big.NewFloat(0)) == 0
}

func (*Float)constant(){}

func (self *Float)FloatValue()*big.Float{
	return self.value
}

func NewNumber(t NumberType, value *big.Float)Number{
	switch tt := t.(type) {
	case IntType:
		if !value.IsInt(){
			panic("unreachable")
		}
		v, _ := value.Int(nil)
		return NewInt(tt, v)
	case FloatType:
		return NewFloat(tt, value)
	default:
		panic("unreachable")
	}
}

// EmptyArray 空数组
type EmptyArray struct {
	t ArrayType
}

func NewEmptyArray(t ArrayType)*EmptyArray{
	return &EmptyArray{t: t}
}

func (self *EmptyArray) Name()string{
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

func (self *EmptyStruct) Name()string{
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

func (self *EmptyFunc) Name()string{
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

func (self *EmptyPtr) Name()string{
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

func NewString(ctx *Context, s string)*Array{
	elems := lo.Map([]byte(s), func(item byte, _ int) Const {
		return NewInt(ctx.U8(), big.NewInt(int64(item)))
	})
	return NewArray(ctx.NewArrayType(uint(len(elems)+1), ctx.U8()), append(elems, NewInt(ctx.U8(), big.NewInt(0)))...).(*Array)
}

func (self *Array) Name()string{
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("[%s]", strings.Join(elems, ","))
}

func (self *Array) Type()Type{
	return self.t
}

func (self *Array) IsZero()bool{
	return false
}

func (self *Array) Elems()[]Const{
	return self.elems
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

func (self *Struct) Name()string{
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("{%s}", strings.Join(elems, ","))
}

func (self *Struct) Type()Type{
	return self.t
}

func (self *Struct) IsZero()bool{
	return false
}

func (self *Struct) Elems()[]Const{
	return self.elems
}

func (*Struct)constant(){}

// ConstArrayIndex 数组索引
type ConstArrayIndex struct {
	i uint
	v Const
	index Const
}

func NewArrayIndex(v, index Const)Const{
	if stlbasic.Is[ArrayType](v.Type()){
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[Array](v.Type().(PtrType).Elem()){
	}else{
		panic("unreachable")
	}
	if !stlbasic.Is[UintType](index.Type()){
		panic("unreachable")
	}
	if vc, ok := v.(Array); ok{
		if ic, ok := index.(Uint); ok{
			return vc.Elems()[ic.IntValue().Uint64()]
		}
	}
	return &ConstArrayIndex{
		v: v,
		index: index,
	}
}

func (self *ConstArrayIndex) Name()string{
	return fmt.Sprintf("array %s index %s", self.v.Name(), self.index.Name())
}

func (self *ConstArrayIndex) Type()Type{
	if stlbasic.Is[ArrayType](self.v.Type()){
		return self.v.Type().(ArrayType).Elem()
	}else{
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(ArrayType).Elem())
	}
}

func (self *ConstArrayIndex) IsZero()bool{
	return false
}

func (*ConstArrayIndex)constant(){}

// ConstStructIndex 结构体索引
type ConstStructIndex struct {
	i uint
	v Const
	index uint
}

func NewStructIndex(v Const, index uint)Const{
	var sizeLength uint
	if stlbasic.Is[StructType](v.Type()){
		sizeLength = uint(len(v.Type().(StructType).Elems()))
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[StructType](v.Type().(PtrType).Elem()){
		sizeLength = uint(len(v.Type().(PtrType).Elem().(StructType).Elems()))
	}else{
		panic("unreachable")
	}
	if index >= sizeLength{
		panic("unreachable")
	}
	if vc, ok := v.(Struct); ok{
		return vc.Elems()[index]
	}
	return &ConstStructIndex{
		v: v,
		index: index,
	}
}

func (self *ConstStructIndex) Name()string{
	return fmt.Sprintf("struct %s index %d", self.v.Name(), self.index)
}

func (self *ConstStructIndex) Type()Type{
	if stlbasic.Is[StructType](self.v.Type()){
		return self.v.Type().(StructType).Elems()[self.index]
	}else{
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(StructType).Elems()[self.index])
	}
}

func (self *ConstStructIndex) IsZero()bool{
	return false
}

func (*ConstStructIndex)constant(){}
