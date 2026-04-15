package cir

import "github.com/kkkunny/stl/enum"

type Type interface {
	typ()
}

var Void = &VoidType{}

type VoidType struct{}

func (*VoidType) typ() {}

var (
	SChar  = &IntType{Kind: IntTypeKindEnum.SChar}
	SShort = &IntType{Kind: IntTypeKindEnum.SShort}
	SInt   = &IntType{Kind: IntTypeKindEnum.SInt}
	SLong  = &IntType{Kind: IntTypeKindEnum.SLong}
	SLLong = &IntType{Kind: IntTypeKindEnum.SLLong}
	UChar  = &IntType{Kind: IntTypeKindEnum.UChar}
	UShort = &IntType{Kind: IntTypeKindEnum.UShort}
	UInt   = &IntType{Kind: IntTypeKindEnum.UInt}
	ULong  = &IntType{Kind: IntTypeKindEnum.ULong}
	ULLong = &IntType{Kind: IntTypeKindEnum.ULLong}
)

type IntTypeKind string

var IntTypeKindEnum = enum.New[struct {
	SChar  IntTypeKind `enum:"signed char"`
	SShort IntTypeKind `enum:"signed short"`
	SInt   IntTypeKind `enum:"signed int"`
	SLong  IntTypeKind `enum:"signed long"`
	SLLong IntTypeKind `enum:"signed long long"`
	UChar  IntTypeKind `enum:"unsigned char"`
	UShort IntTypeKind `enum:"unsigned short"`
	UInt   IntTypeKind `enum:"unsigned int"`
	ULong  IntTypeKind `enum:"unsigned long"`
	ULLong IntTypeKind `enum:"unsigned long long"`
}]()

type IntType struct {
	Kind IntTypeKind
}

func (*IntType) typ() {}

var (
	Float   = &FloatType{Kind: FloatTypeKindEnum.Float}
	Double  = &FloatType{Kind: FloatTypeKindEnum.Double}
	LDouble = &FloatType{Kind: FloatTypeKindEnum.LDouble}
)

type FloatTypeKind string

var FloatTypeKindEnum = enum.New[struct {
	Float   FloatTypeKind `enum:"float"`
	Double  FloatTypeKind `enum:"double"`
	LDouble FloatTypeKind `enum:"long double"`
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
