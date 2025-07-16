package values

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/utils"
)

// Ident 标识符
type Ident interface {
	hir.Value
	GetName() (utils.Name, bool)
	Ident()
}
