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
	SChar  IntTypeKind
	SShort IntTypeKind
	SInt   IntTypeKind
	SLong  IntTypeKind
	SLLong IntTypeKind
	UChar  IntTypeKind
	UShort IntTypeKind
	UInt   IntTypeKind
	ULong  IntTypeKind
	ULLong IntTypeKind
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
	Float   FloatTypeKind
	Double  FloatTypeKind
	LDouble FloatTypeKind
}]()

type FloatType struct {
	Kind FloatTypeKind
}

func (*FloatType) typ() {}
