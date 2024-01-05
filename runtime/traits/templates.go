package traits

import "github.com/kkkunny/Sim/runtime/types"

var (
	defaultTemplate = newTrait("Default", map[string]*types.FuncType{
		"default": types.NewFuncType(selfType),
	})
)

