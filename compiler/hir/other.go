package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
)

// Trait 特性
type Trait struct {
	Pkg string
	Name string
	Methods hashmap.HashMap[string, *FuncType]
}

func DefaultTrait(t Type)*Trait{
	return &Trait{Methods: hashmap.NewHashMapWith[string, *FuncType]("default", &FuncType{Ret: t})}
}