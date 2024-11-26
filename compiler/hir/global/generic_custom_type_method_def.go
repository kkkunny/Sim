package global

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type GenericCustomTypeMethodDef struct {
	*OriginMethodDef
	from GenericCustomTypeDef
}

func newGenericCustomTypeMethodDef(origin *OriginMethodDef, from GenericCustomTypeDef) *GenericCustomTypeMethodDef {
	return &GenericCustomTypeMethodDef{
		OriginMethodDef: origin,
		from:            from,
	}
}

func (self *GenericCustomTypeMethodDef) From() CustomTypeDef {
	return self.from
}

func (self *GenericCustomTypeMethodDef) SetBody(b *local.Block) {
	panic("unreachable")
}

func (self *GenericCustomTypeMethodDef) TotalName(genericParamMap hashmap.HashMap[types.VirtualType, hir.Type]) string {
	name := fmt.Sprintf("%s::%s::%s", self.Package(), self.from.TotalName(genericParamMap), self.name)
	if len(self.genericParams) == 0 {
		return name
	}
	args := stlslices.Map(self.genericParams, func(_ int, arg types.GenericParamType) string {
		return types.ReplaceVirtualType(genericParamMap, arg).String()
	})
	return fmt.Sprintf("%s::<%s>", name, strings.Join(args, ","))
}
