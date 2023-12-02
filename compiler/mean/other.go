package mean

import "github.com/kkkunny/stl/container/hashmap"

// Trait 特性
type Trait struct {
	Methods hashmap.HashMap[string, *FuncType]
}
