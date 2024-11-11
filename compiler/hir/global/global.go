package global

import "github.com/kkkunny/Sim/compiler/hir/types"

type Global interface {
	Package() *Package
	setPackage(pkg *Package)
	Public() bool
	setPublic(pub bool)
}

type TypeDef interface {
	Global
	types.Type
	Target() types.Type
	SetTarget(t types.Type)
	Define() TypeDef
}
