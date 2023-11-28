package mean

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/samber/lo"
)

var (
	Empty = &EmptyType{}

	Isize = &SintType{Bits: 0}
	I8    = &SintType{Bits: 8}
	I16   = &SintType{Bits: 16}
	I32   = &SintType{Bits: 32}
	I64   = &SintType{Bits: 64}
	I128  = &SintType{Bits: 128}

	Usize = &UintType{Bits: 0}
	U8    = &UintType{Bits: 8}
	U16   = &UintType{Bits: 16}
	U32   = &UintType{Bits: 32}
	U64   = &UintType{Bits: 64}
	U128  = &UintType{Bits: 128}

	F32 = &FloatType{Bits: 32}
	F64 = &FloatType{Bits: 64}

	Bool = &BoolType{}

	Str = &StringType{}
)

// Type 类型
type Type interface {
	fmt.Stringer
	Equal(dst Type) bool
	AssignableTo(dst Type) bool
}

// TypeIs 类型是否是
func TypeIs[T Type](v Type) bool {
	_, ok := v.(T)
	return ok
}

// EmptyType 空类型
type EmptyType struct{}

func (_ EmptyType) String() string {
	return "void"
}

func (self *EmptyType) Equal(dst Type) bool {
	_, ok := dst.(*EmptyType)
	return ok
}

func (self *EmptyType) AssignableTo(dst Type) bool {
	return false
}

// NumberType 数字型
type NumberType interface {
	Type
	GetBits() uint
}

// IntType 整型
type IntType interface {
	NumberType
	HasSign() bool
}

// SintType 有符号整型
type SintType struct {
	Bits uint
}

func (self SintType) String() string {
	return fmt.Sprintf("i%d", self.Bits)
}

func (self *SintType) Equal(dst Type) bool {
	t, ok := dst.(*SintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *SintType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

func (self *SintType) HasSign() bool {
	return true
}

func (self *SintType) GetBits() uint {
	return self.Bits
}

// UintType 无符号整型
type UintType struct {
	Bits uint
}

func (self UintType) String() string {
	return fmt.Sprintf("u%d", self.Bits)
}

func (self *UintType) Equal(dst Type) bool {
	t, ok := dst.(*UintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *UintType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

func (self *UintType) HasSign() bool {
	return false
}

func (self *UintType) GetBits() uint {
	return self.Bits
}

// FloatType 浮点型
type FloatType struct {
	Bits uint
}

func (self FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) Equal(dst Type) bool {
	t, ok := dst.(*FloatType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *FloatType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

func (self *FloatType) GetBits() uint {
	return self.Bits
}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func (self FuncType) String() string {
	params := lo.Map(self.Params, func(item Type, _ int) string {
		return item.String()
	})
	ret := stlbasic.Ternary(self.Ret.Equal(Empty), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) Equal(dst Type) bool {
	t, ok := dst.(*FuncType)
	if !ok || len(t.Params) != len(self.Params) {
		return false
	}
	for i, p := range self.Params {
		if !p.Equal(t.Params[i]) {
			return false
		}
	}
	return true
}

func (self *FuncType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// BoolType 布尔型
type BoolType struct{}

func (_ BoolType) String() string {
	return "bool"
}

func (self *BoolType) Equal(dst Type) bool {
	_, ok := dst.(*BoolType)
	return ok
}

func (self *BoolType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// ArrayType 数组型
type ArrayType struct {
	Size uint
	Elem Type
}

func (self ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) Equal(dst Type) bool {
	t, ok := dst.(*ArrayType)
	if !ok {
		return false
	}
	return self.Size == t.Size && self.Elem.Equal(t.Elem)
}

func (self *ArrayType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func (self TupleType) String() string {
	elems := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) Equal(dst Type) bool {
	t, ok := dst.(*TupleType)
	if !ok || len(self.Elems) != len(t.Elems) {
		return false
	}
	for i, e := range self.Elems {
		if !e.Equal(t.Elems[i]) {
			return false
		}
	}
	return true
}

func (self *TupleType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

type StructType = StructDef

// StringType 字符串型
type StringType struct{}

func (_ StringType) String() string {
	return "str"
}

func (self *StringType) Equal(dst Type) bool {
	_, ok := dst.(*StringType)
	return ok
}

func (self *StringType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// UnionType 联合类型
type UnionType struct {
	Elems linkedhashmap.LinkedHashMap[string, Type]
}

func (self UnionType) String() string {
	return fmt.Sprintf("<%s>", strings.Join(self.Elems.Keys().ToSlice(), ", "))
}

func (self *UnionType) Equal(dst Type) bool {
	ut, ok := dst.(*UnionType)
	if !ok {
		return false
	}
	for li, ri := self.Elems.Iterator(), ut.Elems.Iterator(); li.Next() && ri.Next(); {
		if !li.Value().Equal(ri.Value()) {
			return false
		}
	}
	return true
}

func (self *UnionType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// GetElemIndex 获取子类型下标
func (self UnionType) GetElemIndex(elem Type) int {
	if !self.Elems.ContainKey(elem.String()) {
		return -1
	}
	var index int
	for iter := self.Elems.Values().Iterator(); iter.Next(); {
		if elem.Equal(iter.Value()) {
			break
		}
		index++
	}
	return index
}

// PtrType 指针类型
type PtrType struct {
	Elem Type
}

func (self PtrType) String() string {
	return "*" + self.Elem.String() + "?"
}

func (self *PtrType) Equal(dst Type) bool {
	pt, ok := dst.(*PtrType)
	if !ok {
		return false
	}
	return self.Elem.Equal(pt.Elem)
}

func (self *PtrType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// RefType 引用类型
type RefType struct {
	Elem Type
}

func (self RefType) String() string {
	return "*" + self.Elem.String()
}

func (self *RefType) Equal(dst Type) bool {
	pt, ok := dst.(*RefType)
	if !ok {
		return false
	}
	return self.Elem.Equal(pt.Elem)
}

func (self *RefType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return self.ToPtrType().AssignableTo(dst)
}

func (self *RefType) ToPtrType() *PtrType {
	return &PtrType{Elem: self.Elem}
}
