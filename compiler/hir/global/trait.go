package global

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
)

type Trait struct {
	pkgGlobalAttr
	name    string
	methods hashmap.HashMap[string, *FuncDecl]
}

func NewTrait(name string) *Trait {
	return &Trait{
		name:    name,
		methods: hashmap.StdWith[string, *FuncDecl](),
	}
}

func (self *Trait) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *Trait) AddMethod(m *FuncDecl) bool {
	mn, ok := m.GetName()
	if !ok {
		return true
	} else if self.methods.Contain(mn) {
		return false
	}
	self.methods.Set(mn, m)
	return true
}

func (self *Trait) Methods() []*FuncDecl {
	return self.methods.Values()
}

func (self *Trait) FirstMethod() (*FuncDecl, bool) {
	return stlslices.First(self.methods.Values()), !self.methods.Empty()
}
