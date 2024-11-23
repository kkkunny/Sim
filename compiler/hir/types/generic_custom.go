package types

import (
	"github.com/kkkunny/stl/container/hashmap"

	"github.com/kkkunny/Sim/compiler/hir"
)

// GenericCustomType 泛型自定义类型
type GenericCustomType interface {
	CustomType
	Args() []hir.Type
	WithArgs(args []hir.Type) GenericCustomType
	CompileParamMap() hashmap.HashMap[VirtualType, hir.Type]
}
