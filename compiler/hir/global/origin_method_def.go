package global

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type OriginMethodDef struct {
	FuncDef
	from CustomTypeDef
}

func NewOriginMethodDef(from CustomTypeDef, decl *FuncDecl, genericParams []types.GenericParamType, attrs ...FuncAttr) *OriginMethodDef {
	return &OriginMethodDef{
		FuncDef: FuncDef{
			FuncDecl:      decl,
			attrs:         attrs,
			genericParams: genericParams,
		},
		from: from,
	}
}

func (self *OriginMethodDef) From() CustomTypeDef {
	return self.from
}

func (self *OriginMethodDef) SelfParam() (*local.Param, bool) {
	if len(self.params) == 0 {
		return nil, false
	}
	firstParam := stlslices.First(self.params)
	firstParamType := stlval.TernaryAction(types.Is[types.RefType](firstParam.Type(), true), func() hir.Type {
		rt, ok := types.As[types.RefType](firstParam.Type(), true)
		if !ok {
			panic("unreachable")
		}
		return rt.Pointer()
	}, func() hir.Type {
		return firstParam.Type()
	})
	if !firstParamType.Equal(self.from) {
		return nil, false
	}
	return firstParam, true
}

func (self *OriginMethodDef) Static() bool {
	_, ok := self.SelfParam()
	return !ok
}

func (self *OriginMethodDef) SelfParamIsRef() bool {
	selfParam, ok := self.SelfParam()
	if !ok {
		return false
	}
	return types.Is[types.RefType](selfParam.Type(), true)
}

func (self *OriginMethodDef) NotGlobalNamed() {}

func (self *OriginMethodDef) TotalName(genericParamMap hashmap.HashMap[types.VirtualType, hir.Type]) string {
	name := fmt.Sprintf("%s::%s::%s", self.pkg, self.from, self.name)
	if len(self.genericParams) == 0 {
		return name
	}
	args := stlslices.Map(self.genericParams, func(_ int, arg types.GenericParamType) string {
		return types.ReplaceVirtualType(genericParamMap, arg).String()
	})
	return fmt.Sprintf(":%s::<%s>", name, strings.Join(args, ","))
}
