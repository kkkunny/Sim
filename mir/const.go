package mir

import (
	"fmt"
	"strconv"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/samber/lo"
)

// Const 常量
type Const interface {
	Value
	IsZero() bool
	constant()
}

type Number interface {
	Const
	FloatValue() float64
}

type Int interface {
	Number
	IntValue() int64
}

// Sint 有符号整数
type Sint struct {
	t     SintType
	value int64
}

func NewSint(t SintType, value int64) *Sint {
	return &Sint{
		t:     t,
		value: value,
	}
}

func (self *Sint) Name() string {
	return strconv.FormatInt(self.value, 10)
}

func (self *Sint) Type() Type {
	return self.t
}

func (self *Sint) IsZero() bool {
	return self.value == 0
}

func (*Sint) constant() {}

func (self *Sint) FloatValue() float64 {
	return float64(self.value)
}
func (self *Sint) IntValue() int64 {
	return self.value
}

func (self *Sint) ReadRefValues() []Value {
	return nil
}

// Uint 无符号整数
type Uint struct {
	t     UintType
	value uint64
}

func NewUint(t UintType, value uint64) *Uint {
	return &Uint{
		t:     t,
		value: value,
	}
}

func Bool(ctx *Context, v bool) *Uint {
	if v {
		return NewUint(ctx.Bool(), 1)
	} else {
		return NewUint(ctx.Bool(), 0)
	}
}

func (self *Uint) Name() string {
	return strconv.FormatUint(self.value, 10)
}

func (self *Uint) Type() Type {
	return self.t
}

func (self *Uint) IsZero() bool {
	return self.value == 0
}

func (*Uint) constant() {}

func (self *Uint) FloatValue() float64 {
	return float64(self.value)
}
func (self *Uint) IntValue() int64 {
	return int64(self.value)
}

func (self *Uint) ReadRefValues() []Value {
	return nil
}

func NewInt(t IntType, value int64) Int {
	switch tt := t.(type) {
	case SintType:
		return NewSint(tt, value)
	case UintType:
		return NewUint(tt, uint64(value))
	default:
		panic("unreachable")
	}
}

// Float 浮点数
type Float struct {
	t     FloatType
	value float64
}

func NewFloat(t FloatType, value float64) *Float {
	return &Float{
		t:     t,
		value: value,
	}
}

func (self *Float) Name() string {
	return strconv.FormatFloat(self.value, 'E', -1, 64)
}

func (self *Float) Type() Type {
	return self.t
}

func (self *Float) IsZero() bool {
	return self.value == 0
}

func (*Float) constant() {}

func (self *Float) FloatValue() float64 {
	return self.value
}

func (self *Float) ReadRefValues() []Value {
	return nil
}

func NewNumber(t NumberType, value float64) Number {
	switch tt := t.(type) {
	case IntType:
		return NewInt(tt, int64(value))
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

func NewEmptyArray(t ArrayType) *EmptyArray {
	return &EmptyArray{t: t}
}

func (self *EmptyArray) Name() string {
	return "[]"
}

func (self *EmptyArray) Type() Type {
	return self.t
}

func (self *EmptyArray) IsZero() bool {
	return true
}

func (*EmptyArray) constant() {}

func (self *EmptyArray) ReadRefValues() []Value {
	return nil
}

// EmptyStruct 空结构体
type EmptyStruct struct {
	t StructType
}

func NewEmptyStruct(t StructType) *EmptyStruct {
	return &EmptyStruct{t: t}
}

func (self *EmptyStruct) Name() string {
	return "{}"
}

func (self *EmptyStruct) Type() Type {
	return self.t
}

func (self *EmptyStruct) IsZero() bool {
	return true
}

func (*EmptyStruct) constant() {}

func (self *EmptyStruct) ReadRefValues() []Value {
	return nil
}

// EmptyFunc 空函数
type EmptyFunc struct {
	t FuncType
}

func NewEmptyFunc(t FuncType) *EmptyFunc {
	return &EmptyFunc{t: t}
}

func (self *EmptyFunc) Name() string {
	return "nullfunc"
}

func (self *EmptyFunc) Type() Type {
	return self.t
}

func (self *EmptyFunc) IsZero() bool {
	return true
}

func (*EmptyFunc) constant() {}

func (self *EmptyFunc) ReadRefValues() []Value {
	return nil
}

// EmptyPtr 空指针
type EmptyPtr struct {
	t PtrType
}

func NewEmptyPtr(t PtrType) *EmptyPtr {
	return &EmptyPtr{t: t}
}

func (self *EmptyPtr) Name() string {
	return "nullptr"
}

func (self *EmptyPtr) Type() Type {
	return self.t
}

func (self *EmptyPtr) IsZero() bool {
	return true
}

func (*EmptyPtr) constant() {}

func (self *EmptyPtr) ReadRefValues() []Value {
	return nil
}

// NewZero 零值
func NewZero(t Type) Const {
	switch tt := t.(type) {
	case SintType:
		return NewSint(tt, 0)
	case UintType:
		return NewUint(tt, 0)
	case FloatType:
		return NewFloat(tt, 0)
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
	t     ArrayType
	elems []Const
}

func NewArray(t ArrayType, elem ...Const) Const {
	if t.Length() != uint(len(elem)) {
		panic("unreachable")
	}
	for _, e := range elem {
		if !e.Type().Equal(t.Elem()) {
			panic("unreachable")
		}
	}

	zero := stlslices.All(elem, func(_ int, e Const) bool {
		return e.IsZero()
	})
	if zero {
		return NewEmptyArray(t)
	}

	return &Array{
		t:     t,
		elems: elem,
	}
}

func NewString(ctx *Context, s string) Const {
	elems := lo.Map([]byte(s), func(item byte, _ int) Const {
		return NewInt(ctx.U8(), int64(item))
	})
	elems = append(elems, NewInt(ctx.U8(), 0))
	return NewArray(ctx.NewArrayType(uint(len(elems)), ctx.U8()), elems...)
}

func (self *Array) Name() string {
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("[%s]", strings.Join(elems, ","))
}

func (self *Array) Type() Type {
	return self.t
}

func (*Array) IsZero() bool {
	return false
}

func (self *Array) Elems() []Const {
	return self.elems
}

func (*Array) constant() {}

func (self *Array) ReadRefValues() []Value {
	return stlslices.Map(self.elems, func(_ int, e Const) Value {
		return e
	})
}

// Struct 结构体
type Struct struct {
	t     StructType
	elems []Const
}

func NewStruct(t StructType, elem ...Const) Const {
	if len(t.Elems()) != len(elem) {
		panic("unreachable")
	}
	for i, e := range t.Elems() {
		if !e.Equal(elem[i].Type()) {
			panic("unreachable")
		}
	}

	zero := stlslices.All(elem, func(_ int, e Const) bool {
		return e.IsZero()
	})
	if zero {
		return NewEmptyStruct(t)
	}

	return &Struct{
		t:     t,
		elems: elem,
	}
}

func (self *Struct) Name() string {
	elems := lo.Map(self.elems, func(item Const, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("{%s}", strings.Join(elems, ","))
}

func (self *Struct) Type() Type {
	return self.t
}

func (*Struct) IsZero() bool {
	return false
}

func (self *Struct) Elems() []Const {
	return self.elems
}

func (*Struct) constant() {}

func (self *Struct) ReadRefValues() []Value {
	return stlslices.Map(self.elems, func(_ int, e Const) Value {
		return e
	})
}

// ConstArrayIndex 数组索引
type ConstArrayIndex struct {
	i     uint
	v     Const
	index Const
}

func NewArrayIndex(v, index Const) Const {
	if stlbasic.Is[ArrayType](v.Type()) {
	} else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[ArrayType](v.Type().(PtrType).Elem()) {
	} else {
		panic("unreachable")
	}
	if !stlbasic.Is[UintType](index.Type()) {
		panic("unreachable")
	}
	if vc, ok := v.(*Array); ok {
		if ic, ok := index.(*Uint); ok {
			return vc.Elems()[ic.IntValue()]
		}
	}
	return &ConstArrayIndex{
		v:     v,
		index: index,
	}
}

func (self *ConstArrayIndex) Name() string {
	return fmt.Sprintf("array %s index %s", self.v.Name(), self.index.Name())
}

func (self *ConstArrayIndex) Type() Type {
	if self.IsPtr() {
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(ArrayType).Elem())
	}
	return self.v.Type().(ArrayType).Elem()
}

func (*ConstArrayIndex) IsZero() bool {
	return false
}

func (*ConstArrayIndex) constant() {}

func (self *ConstArrayIndex) IsPtr() bool {
	return stlbasic.Is[PtrType](self.v.Type())
}

func (self *ConstArrayIndex) Array() Const {
	return self.v
}

func (self *ConstArrayIndex) Index() Const {
	return self.index
}

func (self *ConstArrayIndex) ReadRefValues() []Value {
	return []Value{self.v}
}

// ConstStructIndex 结构体索引
type ConstStructIndex struct {
	i     uint
	v     Const
	index uint64
}

func NewStructIndex(v Const, index uint64) Const {
	var sizeLength uint64
	if stlbasic.Is[StructType](v.Type()) {
		sizeLength = uint64(len(v.Type().(StructType).Elems()))
	} else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[StructType](v.Type().(PtrType).Elem()) {
		sizeLength = uint64(len(v.Type().(PtrType).Elem().(StructType).Elems()))
	} else {
		panic("unreachable")
	}
	if index >= sizeLength {
		panic("unreachable")
	}
	if vc, ok := v.(*Struct); ok {
		return vc.Elems()[index]
	}
	return &ConstStructIndex{
		v:     v,
		index: index,
	}
}

func (self *ConstStructIndex) Name() string {
	return fmt.Sprintf("struct %s index %d", self.v.Name(), self.index)
}

func (self *ConstStructIndex) Type() Type {
	if self.IsPtr() {
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(StructType).Elems()[self.index])
	}
	return self.v.Type().(StructType).Elems()[self.index]
}

func (*ConstStructIndex) IsZero() bool {
	return false
}

func (*ConstStructIndex) constant() {}

func (self *ConstStructIndex) IsPtr() bool {
	return stlbasic.Is[PtrType](self.v.Type())
}

func (self *ConstStructIndex) Struct() Const {
	return self.v
}

func (self *ConstStructIndex) Index() uint64 {
	return self.index
}

func (self *ConstStructIndex) ReadRefValues() []Value {
	return []Value{self.v}
}
