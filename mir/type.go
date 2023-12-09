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
	Context()Context
	Equal(t Type)bool
	Align()stlos.Size
	Size()stlos.Size
}

// VoidType 空类型
type VoidType struct {
	ctx Context
}

func (self Context) Void()*VoidType{
	return &VoidType{ctx: self}
}

func (self *VoidType) String()string{
	return "void"
}

func (self *VoidType) Context()Context{
	return self.ctx
}

func (self *VoidType) Equal(t Type)bool{
	_, ok := t.(*VoidType)
	return ok
}

func (self *VoidType) Align()stlos.Size{
	return 0
}

func (self *VoidType) Size()stlos.Size{
	return 0
}

// SintType 有符号整型
type SintType struct {
	ctx Context
	size stlos.Size
}

func (self Context) NewSintType(size stlos.Size)*SintType {
	return &SintType{
		ctx: self,
		size: size,
	}
}

func (self Context) I8()*SintType {
	return self.NewSintType(1*stlos.Byte)
}

func (self Context) I16()*SintType {
	return self.NewSintType(2*stlos.Byte)
}

func (self Context) I32()*SintType {
	return self.NewSintType(4*stlos.Byte)
}

func (self Context) I64()*SintType {
	return self.NewSintType(8*stlos.Byte)
}

func (self Context) I128()*SintType {
	return self.NewSintType(16*stlos.Byte)
}

func (self Context) Isize()*SintType {
	// TODO: size
	return self.I64()
}

func (self *SintType) String()string{
	return fmt.Sprintf("i%d", self.size)
}

func (self *SintType) Context()Context{
	return self.ctx
}

func (self *SintType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*SintType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *SintType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *SintType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

// UintType 无符号整型
type UintType struct {
	ctx Context
	size stlos.Size
}

func (self Context) NewUintType(size stlos.Size)*UintType {
	return &UintType{
		ctx: self,
		size: size,
	}
}

func (self Context) U8()*UintType {
	return self.NewUintType(1*stlos.Byte)
}

func (self Context) U16()*UintType {
	return self.NewUintType(2*stlos.Byte)
}

func (self Context) U32()*UintType {
	return self.NewUintType(4*stlos.Byte)
}

func (self Context) U64()*UintType {
	return self.NewUintType(8*stlos.Byte)
}

func (self Context) U128()*UintType {
	return self.NewUintType(16*stlos.Byte)
}

func (self Context) Usize()*UintType {
	// TODO: size
	return self.U64()
}

func (self *UintType) String()string{
	return fmt.Sprintf("u%d", self.size)
}

func (self *UintType) Context()Context{
	return self.ctx
}

func (self *UintType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*UintType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *UintType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *UintType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

// FloatType 浮点型
type FloatType struct {
	ctx Context
	size stlos.Size
}

func (self Context) newFloatType(size stlos.Size)*FloatType{
	return &FloatType{
		ctx: self,
		size: size,
	}
}

func (self Context) F32()*FloatType {
	return self.newFloatType(4*stlos.Byte)
}

func (self Context) F64()*FloatType {
	return self.newFloatType(8*stlos.Byte)
}

func (self *FloatType) String()string{
	return fmt.Sprintf("f%d", self.size)
}

func (self *FloatType) Context()Context{
	return self.ctx
}

func (self *FloatType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*FloatType)
	if !ok{
		return false
	}
	return self.size.Equal(dst.size)
}

func (self *FloatType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *FloatType) Size()stlos.Size{
	// TODO: get size
	return self.size
}

// PtrType 指针类型
type PtrType struct {
	elem Type
}

func (self Context) NewPtrType(elem Type)*PtrType{
	if !elem.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	return &PtrType{elem: elem}
}

func (self *PtrType) String()string{
	return fmt.Sprintf("*%s", self.elem)
}

func (self *PtrType) Context()Context{
	return self.elem.Context()
}

func (self *PtrType) Equal(t Type)bool{
	dst, ok := t.(*PtrType)
	if !ok{
		return false
	}
	return self.elem.Equal(dst.elem)
}

func (self *PtrType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *PtrType) Size()stlos.Size{
	// TODO: get size
	return 8 * stlos.Byte
}

func (self *PtrType) Elem()Type{
	return self.elem
}

// ArrayType 数组类型
type ArrayType struct {
	len uint
	elem Type
}

func (self Context) NewArrayType(len uint, elem Type)*ArrayType{
	if !elem.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	return &ArrayType{
		len: len,
		elem: elem,
	}
}

func (self *ArrayType) String()string{
	return fmt.Sprintf("[%d]%s", self.len, self.elem)
}

func (self *ArrayType) Context()Context{
	return self.elem.Context()
}

func (self *ArrayType) Equal(t Type)bool{
	dst, ok := t.(*ArrayType)
	if !ok{
		return false
	}
	return self.len == dst.len && self.elem.Equal(dst.elem)
}

func (self *ArrayType) Align()stlos.Size{
	// TODO: get align
	return self.elem.Align()
}

func (self *ArrayType) Size()stlos.Size{
	// TODO: get size
	return self.elem.Size() * stlos.Size(self.len)
}

func (self *ArrayType) Length()uint{
	return self.len
}

func (self *ArrayType) Elem()Type{
	return self.elem
}

// StructType 结构体类型
type StructType struct {
	ctx Context
	elems []Type
}

func (self Context) NewStructType(elem ...Type)*StructType{
	for _, e := range elem{
		if !e.Context().Target().Equal(self.Target()){
			panic("unreachable")
		}
	}
	return &StructType{
		ctx: self,
		elems: elem,
	}
}

func (self *StructType) String()string{
	elems := lo.Map(self.elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("{%s}", strings.Join(elems, ","))
}

func (self *StructType) Context()Context{
	return self.ctx
}

func (self *StructType) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*StructType)
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

func (self *StructType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *StructType) Size()stlos.Size{
	// TODO: get size
	return 0
}

func (self *StructType) Elems()[]Type{
	return self.elems
}

// FuncType 函数类型
type FuncType struct {
	ret Type
	params []Type
}

func (self Context) NewFuncType(ret Type, param ...Type)*FuncType{
	if !ret.Context().Target().Equal(self.Target()){
		panic("unreachable")
	}
	for _, p := range param{
		if !p.Context().Target().Equal(self.Target()){
			panic("unreachable")
		}
	}
	return &FuncType{
		ret: ret,
		params: param,
	}
}

func (self *FuncType) String()string{
	params := lo.Map(self.params, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("%s(%s)", self.ret, strings.Join(params, ","))
}

func (self *FuncType) Context()Context{
	return self.ret.Context()
}

func (self *FuncType) Equal(t Type)bool{
	dst, ok := t.(*FuncType)
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

func (self *FuncType) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *FuncType) Size()stlos.Size{
	// TODO: get size
	return 8 * stlos.Byte
}

func (self *FuncType) Ret()Type{
	return self.ret
}

func (self *FuncType) Params()[]Type{
	return self.params
}
