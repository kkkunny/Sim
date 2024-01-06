package traits

import "github.com/kkkunny/Sim/runtime/types"

func NewDefault(selfType types.Type)*Trait{
	return defaultTemplate.copyAndReplaceSelfType(selfType)
}
