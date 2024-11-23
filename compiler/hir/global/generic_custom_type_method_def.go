package global

import (
	"github.com/kkkunny/Sim/compiler/hir/local"
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
