package cir

import "github.com/kkkunny/stl/enum"

type Type interface {
	typ()
}

var Void = &VoidType{}

type VoidType struct{}

func (*VoidType) typ() {}

var (
	I8  = &IntType{Kind: IntTypeKindEnum.I8}
	I16 = &IntType{Kind: IntTypeKindEnum.I16}
	I32 = &IntType{Kind: IntTypeKindEnum.I32}
	I64 = &IntType{Kind: IntTypeKindEnum.I64}
)

type IntTypeKind string

var IntTypeKindEnum = enum.New[struct {
	I8  IntTypeKind `enum:"i8"`
	I16 IntTypeKind `enum:"i16"`
	I32 IntTypeKind `enum:"i32"`
	I64 IntTypeKind `enum:"i64"`
}]()

type IntType struct {
	Kind IntTypeKind
}

func (*IntType) typ() {}

var (
	F32 = &FloatType{Kind: FloatTypeKindEnum.F32}
	F64 = &FloatType{Kind: FloatTypeKindEnum.F64}
)

type FloatTypeKind string

var FloatTypeKindEnum = enum.New[struct {
	F32 FloatTypeKind `enum:"f32"`
	F64 FloatTypeKind `enum:"f64"`
}]()

type FloatType struct {
	Kind FloatTypeKind
}

func (*FloatType) typ() {}

type FuncType struct {
	Return Type
	Params []Type
}

func (*FuncType) typ() {}
