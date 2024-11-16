package types

import "github.com/kkkunny/Sim/compiler/hir"

// CustomType 自定义类型
type CustomType interface {
	hir.Type
	GetName() (string, bool)
	Target() hir.Type
	Custom()
}
