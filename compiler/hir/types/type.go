package types

import (
	"fmt"

	stlcmp "github.com/kkkunny/stl/cmp"
	stlhash "github.com/kkkunny/stl/hash"
)

// Type 类型
type Type interface {
	fmt.Stringer
	stlhash.Hashable
	stlcmp.Equalable[Type]
	EqualWithSelf(dst Type, selfs ...Type) bool
}

// BuildInType 内置类型
type BuildInType interface {
	Type
	BuildIn()
}

// NumType 数字型
type NumType interface {
	BuildInType
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
	BuildInType
	Signed()
}

// CallableType 可调用类型
type CallableType interface {
	BuildInType
	Ret() Type
	Params() []Type
	ToFunc() FuncType
}
