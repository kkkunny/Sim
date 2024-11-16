package types

import "github.com/kkkunny/Sim/compiler/hir"

// AliasType 类型别名
type AliasType interface {
	hir.Type
	GetName() (string, bool)
	Target() hir.Type
	Alias()
}
