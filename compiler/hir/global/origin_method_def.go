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
	return self.FuncDecl.GetSelfParam(self.from)
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
	name := fmt.Sprintf("%s::%s::%s", self.Package(), self.from, self.name)
	if len(self.genericParams) == 0 {
		return name
	}
	args := stlslices.Map(self.genericParams, func(_ int, arg types.GenericParamType) string {
		return types.ReplaceVirtualType(genericParamMap, arg).String()
	})
	return fmt.Sprintf(":%s::<%s>", name, strings.Join(args, ","))
}
