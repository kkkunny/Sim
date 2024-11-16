package values

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Callable 可调用的
type Callable interface {
	hir.Value
	CallableType() types.CallableType
}
