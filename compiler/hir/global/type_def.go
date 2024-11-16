package global

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

type TypeDef interface {
	hir.Global
	hir.Type
	Target() hir.Type
	SetTarget(t hir.Type)
	Define() TypeDef
}
