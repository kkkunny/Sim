package global

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// MethodDef 方法定义
type MethodDef interface {
	hir.Global
	local.CallableDef
	values.Ident
	Attrs() []FuncAttr
	Mutable() bool
	Storable() bool
	From() CustomTypeDef
	SelfParam() (*local.Param, bool)
	Static() bool
	SelfParamIsRef() bool
	NotGlobalNamed()
}
