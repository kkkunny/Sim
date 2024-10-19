package types

import (
	"fmt"

	stlcmp "github.com/kkkunny/stl/cmp"
)

// Type 类型
type Type interface {
	fmt.Stringer
	stlcmp.Equalable[Type]
}

// NumType 数字型
type NumType interface {
	Type
	Number()
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

// IntType 整型
type IntType interface {
	NumType
	Kind() IntTypeKind
}

// SignedType 有符号类型
type SignedType interface {
	Type
	Signed()
}

// CallableType 可调用类型
type CallableType interface {
	Type
	Ret() Type
	Params() []Type
}
