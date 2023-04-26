package mir

import (
	"math/big"
	"strconv"
)

// Constant 常量
type Constant interface {
	Value
	constant()
}

func NewEmpty(t Type) Constant {
	if t.IsNumber() {
		return NewNumber(t, 0)
	} else if t.IsPtr() {
		return NewEmptyPtr(t)
	} else if t.IsFunc() {
		return NewEmptyFunc(t)
	} else if t.IsArray() {
		return NewEmptyArray(t)
	} else if t.IsStruct() {
		return NewEmptyStruct(t)
	} else if t.IsUnion() {
		return NewEmptyUnion(t)
	} else {
		panic("unreachable")
	}
}

// Bool 布尔
type Bool struct {
	Value bool
}

func NewBool(v bool) *Bool {
	return &Bool{
		Value: v,
	}
}
func (self Bool) GetType() Type {
	return NewTypeBool()
}
func (self Bool) constant() {}

func NewNumber(t Type, v float64) Constant {
	if !t.IsNumber() {
		panic("unreachable")
	}
	if t.IsInteger() {
		return NewInt(t, int64(v))
	} else {
		return NewFloat(t, v)
	}
}
func NewNumberFromString(t Type, v string) Constant {
	if !t.IsNumber() {
		panic("unreachable")
	}
	if t.IsInteger() {
		return NewIntFromString(t, v)
	} else {
		return NewFloatFromString(t, v)
	}
}
func NewInt(t Type, v int64) Constant {
	if !t.IsInteger() {
		panic("unreachable")
	}
	if t.IsSint() {
		return NewSint(t, v)
	} else if v > 0 {
		return NewUint(t, uint64(v))
	} else {
		return NewUintFromString(t, strconv.FormatInt(v, 10))
	}
}
func NewIntFromString(t Type, v string) Constant {
	if !t.IsInteger() {
		panic("unreachable")
	}
	if t.IsSint() {
		return NewSintFromString(t, v)
	} else {
		return NewUintFromString(t, v)
	}
}

// Sint 有符号整数
type Sint struct {
	Type  Type
	Value *big.Int
}

func NewSint(t Type, v int64) *Sint {
	if !t.IsSint() {
		panic("unreachable")
	}
	return &Sint{
		Type:  t,
		Value: big.NewInt(v),
	}
}
func NewSintFromString(t Type, v string) *Sint {
	if !t.IsSint() {
		panic("unreachable")
	}
	vv := new(big.Int)
	vv.SetString(v, 10)
	return &Sint{
		Type:  t,
		Value: vv,
	}
}
func (self Sint) GetType() Type {
	return self.Type
}
func (self Sint) constant() {}

// Uint 无符号整数
type Uint struct {
	Type  Type
	Value *big.Int
}

func NewUint(t Type, v uint64) *Uint {
	if !t.IsUint() {
		panic("unreachable")
	}
	return &Uint{
		Type:  t,
		Value: big.NewInt(int64(v)),
	}
}
func NewUintFromString(t Type, v string) *Uint {
	if !t.IsUint() {
		panic("unreachable")
	}
	vv := new(big.Int)
	vv.SetString(v, 10)
	if vv.Cmp(big.NewInt(-1)) != 1 { // if vv < 0
		max := big.NewInt(1)
		max.Lsh(max, t.GetWidth()*8)
		vv.Add(max, vv)
	}
	return &Uint{
		Type:  t,
		Value: vv,
	}
}
func (self Uint) GetType() Type {
	return self.Type
}
func (self Uint) constant() {}

// Float 浮点数
type Float struct {
	Type  Type
	Value *big.Float
}

func NewFloat(t Type, v float64) *Float {
	if !t.IsFloat() {
		panic("unreachable")
	}
	return &Float{
		Type:  t,
		Value: big.NewFloat(v),
	}
}
func NewFloatFromString(t Type, v string) *Float {
	if !t.IsFloat() {
		panic("unreachable")
	}
	vv, _, err := big.ParseFloat(v, 10, big.MaxPrec, big.ToNearestEven)
	if err != nil {
		panic(err)
	}
	return &Float{
		Type:  t,
		Value: vv,
	}
}
func (self Float) GetType() Type {
	return self.Type
}
func (self Float) constant() {}

// EmptyPtr 空指针
type EmptyPtr struct {
	Type Type
}

func NewEmptyPtr(t Type) *EmptyPtr {
	if !t.IsPtr() {
		panic("unreachable")
	}
	return &EmptyPtr{
		Type: t,
	}
}
func (self EmptyPtr) GetType() Type {
	return self.Type
}
func (self EmptyPtr) constant() {}

// EmptyFunc 空函数
type EmptyFunc struct {
	Type Type
}

func NewEmptyFunc(t Type) *EmptyFunc {
	if !t.IsFunc() {
		panic("unreachable")
	}
	return &EmptyFunc{
		Type: t,
	}
}
func (self EmptyFunc) GetType() Type {
	return self.Type
}
func (self EmptyFunc) constant() {}

// EmptyArray 空数组
type EmptyArray struct {
	Type Type
}

func NewEmptyArray(t Type) *EmptyArray {
	if !t.IsArray() {
		panic("unreachable")
	}
	return &EmptyArray{
		Type: t,
	}
}
func (self EmptyArray) GetType() Type {
	return self.Type
}
func (self EmptyArray) constant() {}

// EmptyStruct 空结构体
type EmptyStruct struct {
	Type Type
}

func NewEmptyStruct(t Type) *EmptyStruct {
	if !t.IsStruct() {
		panic("unreachable")
	}
	return &EmptyStruct{
		Type: t,
	}
}
func (self EmptyStruct) GetType() Type {
	return self.Type
}
func (self EmptyStruct) constant() {}

// EmptyUnion 空结构体
type EmptyUnion struct {
	Type Type
}

func NewEmptyUnion(t Type) *EmptyUnion {
	if !t.IsUnion() {
		panic("unreachable")
	}
	return &EmptyUnion{
		Type: t,
	}
}
func (self EmptyUnion) GetType() Type {
	return self.Type
}
func (self EmptyUnion) constant() {}

// Array 数组
type Array struct {
	Elems []Constant
}

func NewArray(elem ...Constant) *Array {
	if len(elem) == 0 {
		panic("unreachable")
	}
	t := elem[0].GetType()
	for _, e := range elem[1:] {
		if !e.GetType().Equal(t) {
			panic("unreachable")
		}
	}
	return &Array{
		Elems: elem,
	}
}
func (self Array) GetType() Type {
	return NewTypeArray(uint(len(self.Elems)), self.Elems[0].GetType())
}
func (self Array) constant() {}

// Struct 结构体
type Struct struct {
	Elems []Constant
}

func NewStruct(elem ...Constant) *Struct {
	if len(elem) == 0 {
		panic("unreachable")
	}
	return &Struct{
		Elems: elem,
	}
}
func (self Struct) GetType() Type {
	elems := make([]Type, len(self.Elems))
	for i, e := range self.Elems {
		elems[i] = e.GetType()
	}
	return NewTypeStruct(elems...)
}
func (self Struct) constant() {}

// ArrayIndexConst 数组索引
type ArrayIndexConst struct {
	From  Constant
	Index uint
}

func NewArrayIndexConst(f Constant, i uint) *ArrayIndexConst {
	if !f.GetType().IsArray() && !(f.GetType().IsPtr() && f.GetType().GetPtr().IsArray()) {
		panic("unreachable")
	}
	return &ArrayIndexConst{
		From:  f,
		Index: i,
	}
}
func (self ArrayIndexConst) GetType() Type {
	ft := self.From.GetType()
	if ft.IsPtr() {
		return NewTypePtr(ft.GetPtr().GetArrayElem())
	}
	return ft.GetArrayElem()
}
func (self ArrayIndexConst) constant() {}
func (self ArrayIndexConst) IsPtr() bool {
	if self.From.GetType().IsPtr() {
		return true
	}
	return false
}
