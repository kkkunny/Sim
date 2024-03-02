package types

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"
)

var (
	TypeNoThing  = new(NoThingType)
	TypeNoReturn = new(NoReturnType)
)

const (
	_ uint64 = iota
	emptyTypeHash
	noReturnTypeHash
)

// Type 类型描述
type Type interface {
	fmt.Stringer
	stlbasic.Hashable
	stlbasic.Comparable[Type]
}

// NoThingType 无返回值类型
type NoThingType struct{}

func (*NoThingType) String() string {
	return "void"
}

func (self *NoThingType) Hash() uint64 {
	return emptyTypeHash
}

func (self *NoThingType) Equal(dst Type) bool {
	return stlbasic.Is[*NoThingType](dst)
}

// NoReturnType 无返回类型
type NoReturnType struct{}

func (*NoReturnType) String() string {
	return "X"
}

func (self *NoReturnType) Hash() uint64 {
	return noReturnTypeHash
}

func (self *NoReturnType) Equal(dst Type) bool {
	return stlbasic.Is[*NoReturnType](dst)
}

// SintType 有符号整型
type SintType struct {
	Bits uint8
}

func NewSintType(bits uint8) *SintType {
	return &SintType{Bits: bits}
}

func (self *SintType) String() string {
	return fmt.Sprintf("i%d", self.Bits)
}

func (self *SintType) Hash() uint64 {
	return uint64(self.Bits)
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
	Bits uint8
}

func NewUintType(bits uint8) *UintType {
	return &UintType{Bits: bits}
}

func (self *UintType) String() string {
	return fmt.Sprintf("u%d", self.Bits)
}

func (self *UintType) Hash() uint64 {
	return uint64(self.Bits)
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
	Bits uint8
}

func NewFloatType(bits uint8) *FloatType {
	return &FloatType{Bits: bits}
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) Hash() uint64 {
	return uint64(self.Bits)
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
	Mut  bool
	Elem Type
}

func NewRefType(mut bool, elem Type) *RefType {
	return &RefType{
		Mut:  mut,
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

func NewFuncType(ret Type, params ...Type) *FuncType {
	return &FuncType{
		Ret:    ret,
		Params: params,
	}
}

func (self *FuncType) String() string {
	params := stlslices.Map(self.Params, func(_ int, item Type) string {
		return item.String()
	})
	ret := stlbasic.Ternary(self.Ret.Equal(TypeNoThing), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
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

func NewArrayType(size uint64, elem Type) *ArrayType {
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

func NewTupleType(elems ...Type) *TupleType {
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

type Field struct {
	Type Type
	Name string
}

func NewField(t Type, name string) Field {
	return Field{
		Type: t,
		Name: name,
	}
}

type Method struct {
	Type *FuncType
	Name string
}

func NewMethod(t *FuncType, name string) Method {
	return Method{
		Type: t,
		Name: name,
	}
}

// StructType 结构体类型
type StructType struct {
	Pkg    string
	Fields []Field
}

func NewStructType(pkg string, fields ...Field) *StructType {
	return &StructType{
		Pkg:    pkg,
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
	for i, f := range self.Fields {
		if f.Name != dt.Fields[i].Name || !f.Type.Equal(dt.Fields[i].Type) {
			return false
		}
	}
	return true
}

// CustomType 自定义类型
type CustomType struct {
	Pkg     string
	Name    string
	Target  Type
	Methods []Method
}

func NewCustomType(pkg, name string, target Type, methods []Method) *CustomType {
	return &CustomType{
		Pkg:     pkg,
		Name:    name,
		Target:  target,
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

// LambdaType 匿名函数类型
type LambdaType struct {
	Ret    Type
	Params []Type
}

func NewLambdaType(ret Type, params ...Type) *LambdaType {
	return &LambdaType{
		Ret:    ret,
		Params: params,
	}
}

func (self *LambdaType) String() string {
	params := stlslices.Map(self.Params, func(_ int, item Type) string {
		return item.String()
	})
	ret := stlbasic.Ternary(self.Ret.Equal(TypeNoThing), "", self.Ret.String())
	return fmt.Sprintf("(%s)->%s", strings.Join(params, ", "), ret)
}

func (self *LambdaType) Hash() uint64 {
	return self.Ret.Hash() & stlbasic.Hash(self.Params)
}

func (self *LambdaType) Equal(dst Type) bool {
	dt, ok := dst.(*LambdaType)
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

type EnumField struct {
	Name string
}

func NewEnumField(name string) EnumField {
	return EnumField{
		Name: name,
	}
}

// EnumType 枚举类型
type EnumType struct {
	Pkg    string
	Fields []EnumField
}

func NewEnumType(pkg string, fields ...EnumField) *EnumType {
	return &EnumType{
		Pkg:    pkg,
		Fields: fields,
	}
}

func (self *EnumType) String() string {
	return fmt.Sprintf("<%s>", stlslices.Map(self.Fields, func(_ int, e EnumField) string {
		return e.Name
	}))
}

func (self *EnumType) Hash() uint64 {
	return stlbasic.Hash(stlslices.Map(self.Fields, func(_ int, e EnumField) string {
		return e.Name
	}))
}

func (self *EnumType) Equal(dst Type) bool {
	dt, ok := dst.(*EnumType)
	if !ok || self.Pkg != dt.Pkg || len(self.Fields) != len(dt.Fields) {
		return false
	}
	for i, f := range self.Fields {
		if f.Name != dt.Fields[i].Name {
			return false
		}
	}
	return true
}
