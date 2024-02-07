package types

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"
)

var (
	TypeEmpty = new(EmptyType)
	TypeBool = new(BoolType)
	TypeStr = new(StringType)

	TypeI8 = NewSintType(8)
	TypeI16 = NewSintType(16)
	TypeI32 = NewSintType(32)
	TypeI64 = NewSintType(64)
	TypeIsize = TypeI64

	TypeU8 = NewUintType(8)
	TypeU16 = NewUintType(16)
	TypeU32 = NewUintType(32)
	TypeU64 = NewUintType(64)
	TypeUsize = TypeU64

	TypeF32 = NewFloatType(32)
	TypeF64 = NewFloatType(64)
)

const (
	_ uint64 = iota
	emptyTypeHash
	boolTypeHash
	stringTypeHash
)

// Type 类型描述
type Type interface {
	fmt.Stringer
	stlbasic.Hashable
	stlbasic.Comparable[Type]
}

// EmptyType 空类型
type EmptyType struct{}

func (*EmptyType) String() string {
	return ""
}

func (self *EmptyType) Hash() uint64 {
	return emptyTypeHash
}

func (self *EmptyType) Equal(dst Type) bool {
	return stlbasic.Is[*EmptyType](dst)
}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) Hash() uint64 {
	return boolTypeHash
}

func (self *BoolType) Equal(dst Type) bool {
	return stlbasic.Is[*BoolType](dst)
}

// StringType 字符串型
type StringType struct{}

func (*StringType) String() string {
	return "str"
}

func (self *StringType) Hash() uint64 {
	return stringTypeHash
}

func (self *StringType) Equal(dst Type) bool {
	return stlbasic.Is[*StringType](dst)
}

// SintType 有符号整型
type SintType struct {
	Bits uint64
}

func NewSintType(bits uint64)*SintType{
	return &SintType{Bits: bits}
}

func (self *SintType) String() string {
	return fmt.Sprintf("i%d", self.Bits)
}

func (self *SintType) Hash() uint64 {
	return self.Bits
}

func (self *SintType) Equal(dst Type) bool {
	dt, ok := dst.(*SintType)
	if !ok {
		return false
	}
	return self.Bits == dt.Bits
}

// UintType 无符号整型
type UintType struct {
	Bits uint64
}

func NewUintType(bits uint64)*UintType{
	return &UintType{Bits: bits}
}

func (self *UintType) String() string {
	return fmt.Sprintf("u%d", self.Bits)
}

func (self *UintType) Hash() uint64 {
	return self.Bits
}

func (self *UintType) Equal(dst Type) bool {
	dt, ok := dst.(*UintType)
	if !ok {
		return false
	}
	return self.Bits == dt.Bits
}

// FloatType 浮点型
type FloatType struct {
	Bits uint64
}

func NewFloatType(bits uint64)*FloatType{
	return &FloatType{Bits: bits}
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) Hash() uint64 {
	return self.Bits
}

func (self *FloatType) Equal(dst Type) bool {
	dt, ok := dst.(*FloatType)
	if !ok {
		return false
	}
	return self.Bits == dt.Bits
}

// RefType 引用类型
type RefType struct {
	Mut bool
	Elem Type
}

func NewRefType(mut bool, elem Type)*RefType{
	return &RefType{
		Mut: mut,
		Elem: elem,
	}
}

func (self *RefType) String() string {
	return stlbasic.Ternary(self.Mut, "&mut ", "&") + self.Elem.String()
}

func (self *RefType) Hash() uint64 {
	return self.Elem.Hash()<<2 + 1
}

func (self *RefType) Equal(dst Type) bool {
	dt, ok := dst.(*RefType)
	if !ok {
		return false
	}
	return self.Mut == dt.Mut && self.Elem.Equal(dt.Elem)
}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func NewFuncType(ret Type, params ...Type)*FuncType{
	return &FuncType{
		Ret: ret,
		Params: params,
	}
}

func (self *FuncType) String() string {
	params := lo.Map(self.Params, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), self.Ret)
}

func (self *FuncType) Hash() uint64 {
	return self.Ret.Hash() & stlbasic.Hash(self.Params)
}

func (self *FuncType) Equal(dst Type) bool {
	dt, ok := dst.(*FuncType)
	if !ok || len(self.Params) != len(dt.Params) || !self.Ret.Equal(dt.Ret) {
		return false
	}
	for i, p := range self.Params {
		if !p.Equal(dt.Params[i]) {
			return false
		}
	}
	return true
}

// ArrayType 数组型
type ArrayType struct {
	Size uint64
	Elem Type
}

func NewArrayType(size uint64, elem Type)*ArrayType{
	return &ArrayType{
		Size: size,
		Elem: elem,
	}
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) Hash() uint64 {
	return self.Elem.Hash() * self.Size
}

func (self *ArrayType) Equal(dst Type) bool {
	dt, ok := dst.(*ArrayType)
	if !ok || self.Size != dt.Size || !self.Elem.Equal(dt.Elem) {
		return false
	}
	return true
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func NewTupleType(elems ...Type)*TupleType{
	return &TupleType{Elems: elems}
}

func (self *TupleType) String() string {
	elems := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) Hash() (h uint64) {
	return stlbasic.Hash(self.Elems)
}

func (self *TupleType) Equal(dst Type) bool {
	dt, ok := dst.(*TupleType)
	if !ok || len(self.Elems) != len(dt.Elems) {
		return false
	}
	for i, e := range self.Elems {
		if !e.Equal(dt.Elems[i]) {
			return false
		}
	}
	return true
}

// UnionType 联合类型
type UnionType struct {
	Elems []Type
}

func NewUnionType(elems ...Type)*UnionType{
	elems = stlslices.FlatMap(elems, func(_ int, e Type) []Type {
		if ue, ok := e.(*UnionType); ok{
			return ue.Elems
		}else{
			return []Type{e}
		}
	})
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].String() < elems[j].String()
	})

	flatElems := make([]Type, 0, len(elems))
loop:
	for _, e := range elems{
		for _, fe := range flatElems{
			if e.Equal(fe){
				continue loop
			}
		}
		flatElems = append(flatElems, e)
	}

	return &UnionType{Elems: flatElems}
}

func (self *UnionType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string {return e.String()})
	return fmt.Sprintf("<%s>", strings.Join(elems, ", "))
}

func (self *UnionType) Hash() uint64 {
	return slices.Max(stlslices.Map(self.Elems, func(_ int, e Type) uint64 {
		return e.Hash()
	})) | 16
}

func (self *UnionType) Equal(dst Type) bool {
	dt, ok := dst.(*UnionType)
	if !ok {
		return false
	}
	for i, e := range self.Elems {
		if !e.Equal(dt.Elems[i]) {
			return false
		}
	}
	return true
}

func (self *UnionType) IndexElem(t Type)int{
	for i, elem := range self.Elems{
		if elem.Equal(t){
			return i
		}
	}
	return -1
}

func (self *UnionType) Elem(i uint)Type{
	return self.Elems[i]
}

func (self *UnionType) Contain(t Type)bool{
	if ut, ok := t.(*UnionType); ok{
		return stlslices.All[Type](ut.Elems, func(_ int, e Type) bool {return self.Contain(e)})
	}else{
		return stlslices.Contain(self.Elems, t)
	}
}

type Field struct {
	Type Type
	Name string
}

func NewField(t Type, name string)Field{
	return Field{
		Type: t,
		Name: name,
	}
}

type Method struct {
	Type *FuncType
	Name string
}

func NewMethod(t *FuncType, name string)Method{
	return Method{
		Type: t,
		Name: name,
	}
}

// StructType 结构体类型
type StructType struct {
	Pkg string
	Fields []Field
}

func NewStructType(pkg string, fields ...Field)*StructType {
	return &StructType{
		Pkg: pkg,
		Fields: fields,
	}
}

func (self *StructType) String() string {
	return fmt.Sprintf("{%s}", stlslices.Map(self.Fields, func(_ int, e Field) string {
		return fmt.Sprintf("%s: %s", e.Name, e.Type)
	}))
}

func (self *StructType) Hash() uint64 {
	return stlbasic.Hash(stlslices.Map(self.Fields, func(_ int, e Field) string {
		return fmt.Sprintf("%s: %s", e.Name, e.Type)
	}))
}

func (self *StructType) Equal(dst Type) bool {
	dt, ok := dst.(*StructType)
	if !ok || self.Pkg != dt.Pkg || len(self.Fields) != len(dt.Fields) {
		return false
	}
	for i, f := range self.Fields{
		if f.Name != dt.Fields[i].Name || !f.Type.Equal(dt.Fields[i].Type){
			return false
		}
	}
	return true
}

// CustomType 自定义类型
type CustomType struct {
	Pkg    string
	Name   string
	Target Type
	Methods []Method
}

func NewCustomType(pkg, name string, target Type, methods []Method)*CustomType {
	return &CustomType{
		Pkg: pkg,
		Name: name,
		Target: target,
		Methods: methods,
	}
}

func (self *CustomType) String() string {
	return fmt.Sprintf("%s::%s", self.Pkg, self.Name)
}

func (self *CustomType) Hash() uint64 {
	return stlbasic.Hash(self.String())
}

func (self *CustomType) Equal(dst Type) bool {
	dt, ok := dst.(*CustomType)
	if !ok {
		return false
	}
	return self.Pkg == dt.Pkg && self.Name == dt.Name
}
