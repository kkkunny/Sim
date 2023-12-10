package mir

import (
	"fmt"
	"strings"

	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"
)

// Type 类型
type Type interface {
	fmt.Stringer
	Context()*Context
	Equal(t Type)bool
	Align()stlos.Size
	Size()stlos.Size
}

// VoidType 空类型
type VoidType interface {
	Type
	void()
}

type voidType struct {
	ctx *Context
}

func (self *Context) Void()VoidType{
	return &voidType{ctx: self}
}

func (self *voidType) String()string{
	return "void"
}

func (self *voidType) Context()*Context{
	return self.ctx
}

func (self *voidType) Equal(t Type)bool{
	_, ok := t.(*voidType)
	return ok
}

func (self *voidType) Align()stlos.Size{
	return 0
}

func (self *voidType) Size()stlos.Size{
	return 0
}

func (self *voidType) void(){}

// NumberType 数字类型
type NumberType interface {
	Type
	number()
}

// IntType 整型
type IntType interface {
	NumberType
	integer()
}

// SintType 有符号整型
type SintType interface {
	IntType
	sint()
}

type sintType struct {
	ctx *Context
	size stlos.Size
}

func (self *Context) NewSintType(size stlos.Size)SintType {
	return &sintType{
		ctx: self,
		size: size,
	}
}

func (self *Context) I8()SintType {
	return self.NewSintType(1*stlos.Byte)
}

func (self *Context) I16()SintType {
	return self.NewSintType(2*stlos.Byte)
}

func (self *Context) I32()SintType {
	return self.NewSintType(4*stlos.Byte)
}

func (self *Context) I64()SintType {
	return self.NewSintType(8*stlos.Byte)
}

func (self *Context) I128()SintType {
	return self.NewSintType(16*stlos.Byte)
}

func (self *Context) Isize()SintType {
	// TODO: size
	return self.I64()
}

func (self *sintType) String()string{
	return fmt.Sprintf("i%d", self.size)
}

func (self *sintType) Context()*Context{
	return self.ctx
}

func (self *sintType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*sintType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *sintType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *sintType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

func (*sintType) number(){}
func (*sintType) integer() {}
func (*sintType) sint() {}

// UintType 无符号整型
type UintType interface {
	IntType
	uint()
}

// uintType 无符号整型
type uintType struct {
	ctx *Context
	size stlos.Size
}

func (self *Context) NewUintType(size stlos.Size)UintType {
	return &uintType{
		ctx: self,
		size: size,
	}
}

func (self *Context) Bool()UintType {
	return self.NewUintType(1*stlos.Bit)
}

func (self *Context) U8()UintType {
	return self.NewUintType(1*stlos.Byte)
}

func (self *Context) U16()UintType {
	return self.NewUintType(2*stlos.Byte)
}

func (self *Context) U32()UintType {
	return self.NewUintType(4*stlos.Byte)
}

func (self *Context) U64()UintType {
	return self.NewUintType(8*stlos.Byte)
}

func (self *Context) U128()UintType {
	return self.NewUintType(16*stlos.Byte)
}

func (self *Context) Usize()UintType {
	// TODO: size
	return self.U64()
}

func (self *uintType) String()string{
	return fmt.Sprintf("u%d", self.size)
}

func (self *uintType) Context()*Context{
	return self.ctx
}

func (self *uintType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*uintType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *uintType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *uintType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

func (*uintType) number(){}
func (*uintType) integer() {}
func (*uintType) uint(){}

// FloatType 浮点型
type FloatType interface {
	NumberType
	float()
}

type floatType struct {
	ctx *Context
	size stlos.Size
}

func (self *Context) newFloatType(size stlos.Size)FloatType {
	return &floatType{
		ctx: self,
		size: size,
	}
}

func (self *Context) F32()FloatType {
	return self.newFloatType(4*stlos.Byte)
}

func (self *Context) F64()FloatType {
	return self.newFloatType(8*stlos.Byte)
}

func (self *floatType) String()string{
	return fmt.Sprintf("f%d", self.size)
}

func (self *floatType) Context()*Context{
	return self.ctx
}

func (self *floatType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*floatType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *floatType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *floatType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

func (*floatType) number() {}
func (*floatType) float() {}

type GenericPtrType interface {
	Type
	ptr()
}

// PtrType 指针类型
type PtrType interface {
	GenericPtrType
	Elem()Type
}

type ptrType struct {
	elem Type
}

func (self *Context) NewPtrType(elem Type)PtrType {
	if !elem.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	return &ptrType{elem: elem}
}

func (self *ptrType) String()string{
	return fmt.Sprintf("*%s", self.elem)
}

func (self *ptrType) Context()*Context{
	return self.elem.Context()
}

func (self *ptrType) Equal(t Type)bool{
	dst, ok := t.(*ptrType)
	if !ok{
		return false
	}
	return self.elem.Equal(dst.elem)
}

func (self *ptrType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *ptrType) Size()stlos.Size{
	// TODO: get size
	return 8 * stlos.Byte
}

func (self *ptrType) Elem()Type{
	return self.elem
}

func (self *ptrType) ptr(){}

// ArrayType 数组类型
type ArrayType interface {
	Type
	Length()uint
	Elem()Type
	array()
}

type arrayType struct {
	len uint
	elem Type
}

func (self *Context) NewArrayType(len uint, elem Type)ArrayType {
	if !elem.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	return &arrayType{
		len: len,
		elem: elem,
	}
}

func (self *arrayType) String()string{
	return fmt.Sprintf("[%d]%s", self.len, self.elem)
}

func (self *arrayType) Context()*Context{
	return self.elem.Context()
}

func (self *arrayType) Equal(t Type)bool{
	dst, ok := t.(*arrayType)
	if !ok{
		return false
	}
	return self.len == dst.len && self.elem.Equal(dst.elem)
}

func (self *arrayType) Align()stlos.Size{
	// TODO: get align
	return self.elem.Align()
}

func (self *arrayType) Size()stlos.Size{
	// TODO: get size
	return self.elem.Size() * stlos.Size(self.len)
}

func (self *arrayType) Length()uint{
	return self.len
}

func (self *arrayType) Elem()Type{
	return self.elem
}

func (self *arrayType) array(){}

// StructType 结构体类型
type StructType interface {
	Type
	SetElems(elem ...Type)
	Elems()[]Type
}

// 无名字结构体类型
type unnamedStructType struct {
	ctx *Context
	elems []Type
}

func (self *Context) NewStructType(elem ...Type) StructType {
	for _, e := range elem{
		if !e.Context().Target().Equal(self.Target()){
			panic("unreachable")
		}
	}
	return &unnamedStructType{
		ctx: self,
		elems: elem,
	}
}

func (self *unnamedStructType) String()string{
	elems := lo.Map(self.elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("{%s}", strings.Join(elems, ","))
}

func (self *unnamedStructType) Context()*Context{
	return self.ctx
}

func (self *unnamedStructType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*unnamedStructType)
	if !ok || len(self.elems) != len(dst.elems){
		return false
	}
	for i, e := range self.elems{
		if !e.Equal(dst.elems[i]){
			return false
		}
	}
	return true
}

func (self *unnamedStructType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *unnamedStructType) Size()stlos.Size{
	// TODO: get size
	return 0
}

func (self *unnamedStructType) SetElems(elem ...Type){
	self.elems = elem
}

func (self *unnamedStructType) Elems()[]Type{
	return self.elems
}

// FuncType 函数类型
type FuncType interface {
	Type
	Ret()Type
	Params()[]Type
}

type funcType struct {
	ret Type
	params []Type
}

func (self *Context) NewFuncType(ret Type, param ...Type)FuncType {
	if !ret.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	for _, p := range param{
		if !p.Context().Target().Equal(self.Target()){
			panic("unreachable")
		}
	}
	return &funcType{
		ret: ret,
		params: param,
	}
}

func (self *funcType) String()string{
	params := lo.Map(self.params, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("%s(%s)", self.ret, strings.Join(params, ","))
}

func (self *funcType) Context()*Context{
	return self.ret.Context()
}

func (self *funcType) Equal(t Type)bool{
	dst, ok := t.(*funcType)
	if !ok || !self.ret.Equal(dst.ret) || len(self.params) != len(dst.params){
		return false
	}
	for i, p := range self.params{
		if !p.Equal(dst.params[i]){
			return false
		}
	}
	return true
}

func (self *funcType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *funcType) Size()stlos.Size{
	// TODO: get size
	return 8 * stlos.Byte
}

func (self *funcType) Ret()Type{
	return self.ret
}

func (self *funcType) Params()[]Type{
	return self.params
}

func (self *funcType) ptr(){}
