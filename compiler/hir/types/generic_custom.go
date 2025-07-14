package types

import (
	"github.com/kkkunny/stl/container/hashmap"

	"github.com/kkkunny/Sim/compiler/hir"
)

// GenericCustomType 泛型自定义类型
type GenericCustomType interface {
	CustomType
	GenericArgs() []hir.Type
	WithGenericArgs(args []hir.Type) GenericCustomType
	GenericParamMap() hashmap.HashMap[VirtualType, hir.Type]
	TotalName(genericParamMap hashmap.HashMap[VirtualType, hir.Type]) string
}
