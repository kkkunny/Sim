package global

import (
	"github.com/kkkunny/stl/container/hashmap"
)

type Trait struct {
	pkgGlobalAttr
	name    string
	Methods hashmap.HashMap[string, *FuncDecl]
}

func NewTrait(name string, methods hashmap.HashMap[string, *FuncDecl]) *Trait {
	return &Trait{
		name:    name,
		Methods: methods,
	}
}

func (self *Trait) Name() string {
	return self.name
}
