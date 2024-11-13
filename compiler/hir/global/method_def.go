package global

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// MethodDef 方法定义
type MethodDef struct {
	FuncDef
	from CustomTypeDef
}

func NewMethodDef(from CustomTypeDef, decl *FuncDecl, attrs ...FuncAttr) *MethodDef {
	return &MethodDef{
		FuncDef: FuncDef{
			FuncDecl: decl,
			attrs:    attrs,
		},
		from: from,
	}
}

func (self *MethodDef) From() CustomTypeDef {
	return self.from
}

func (self *MethodDef) SelfParam() (*local.Param, bool) {
	if len(self.params) == 0 {
		return nil, false
	}
	firstParam := stlslices.First(self.params)
	firstParamType := stlval.TernaryAction(types.Is[types.RefType](firstParam.Type(), true), func() types.Type {
		rt, ok := types.As[types.RefType](firstParam.Type(), true)
		if !ok {
			panic("unreachable")
		}
		return rt.Pointer()
	}, func() types.Type {
		return firstParam.Type()
	})
	if !firstParamType.Equal(self.from) {
		return nil, false
	}
	return firstParam, true
}

func (self *MethodDef) Static() bool {
	_, ok := self.SelfParam()
	return !ok
}

func (self *MethodDef) SelfParamIsRef() bool {
	selfParam, ok := self.SelfParam()
	if !ok {
		return false
	}
	return types.Is[types.RefType](selfParam.Type(), true)
}
