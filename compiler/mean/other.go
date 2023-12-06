package mean

import (
	"github.com/kkkunny/stl/container/hashmap"
)

// Trait 特性
type Trait struct {
	Methods hashmap.HashMap[string, *FuncType]
}

func DefaultTrait(t Type)*Trait{
	return &Trait{Methods: hashmap.NewHashMapWith[string, *FuncType]("default", &FuncType{Ret: t})}
}
