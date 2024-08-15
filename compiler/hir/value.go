package hir

import "github.com/kkkunny/Sim/compiler/hir/types"

type Value interface {
	Type() types.Type
	Mutable() bool
}
