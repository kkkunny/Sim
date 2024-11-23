package types

import "github.com/kkkunny/Sim/compiler/hir"

type wrapper interface {
	Wrap(inner hir.Type) hir.BuildInType
}

func wrap[To hir.Type](from hir.Type, inner To) To {
	return from.(wrapper).Wrap(inner).(To)
}
