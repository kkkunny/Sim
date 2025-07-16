package types

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/utils"
)

// NumType 数字型
type NumType interface {
	hir.BuildInType
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
	hir.BuildInType
	Signed()
}

// CallableType 可调用类型
type CallableType interface {
	hir.BuildInType
	Ret() hir.Type
	Params() []hir.Type
	ToFunc() FuncType
}

// TypeDef 定义的类型
type TypeDef interface {
	hir.Global
	hir.Type
	Target() hir.Type
	SetTarget(t hir.Type)
	GetName() (utils.Name, bool)
	Wrap(inner hir.Type) hir.BuildInType
}

// VirtualType 占位类型
type VirtualType interface {
	hir.Type
	virtual()
}
