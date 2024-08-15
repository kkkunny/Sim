package types

import (
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"
)

// Type 类型
type Type interface {
	fmt.Stringer
	stlbasic.Comparable[Type]
}

// 数字型
type numType interface {
	Type
	num()
}

// IntTypeKind 整型类型
type IntTypeKind uint8

const (
	IntTypeKindSize IntTypeKind = 0
	IntTypeKindByte IntTypeKind = 1 << (iota - 1)
	IntTypeKindShort
	IntTypeKindInt
	IntTypeKindLong
)

// 整型
type intType interface {
	numType
	Kind() IntTypeKind
}

// 有符号类型
type signedType interface {
	Type
	signed()
}
