package types

import "github.com/kkkunny/Sim/compiler/hir"

// BuildInType 内置类型
type BuildInType interface {
	hir.Type
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
	Ret() hir.Type
	Params() []hir.Type
	ToFunc() FuncType
}
