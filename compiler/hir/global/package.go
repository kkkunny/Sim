package global

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlhash "github.com/kkkunny/stl/hash"
	stlos "github.com/kkkunny/stl/os"
)

type pkgGlobalAttr struct {
	pkg *Package
	pub bool
}

func (self pkgGlobalAttr) Package() *Package {
	return self.pkg
}

func (self pkgGlobalAttr) Public() bool {
	return self.pub
}

type Package struct {
	path    stlos.FilePath
	globals linkedlist.LinkedList[Global]

	externs hashmap.HashMap[string, []*Package]
	idents  hashmap.HashMap[string, any]
}

func NewPackage() *Package {
	return &Package{
		globals: linkedlist.NewLinkedList[Global](),
		externs: hashmap.StdWith[string, []*Package](),
	}
}

func (self *Package) String() string {
	// TODO
	panic("unreachable")
}

func (self *Package) Equal(dst *Package) bool {
	if self == dst {
		return true
	}
	return *self == *dst
}

func (self *Package) Hash() uint64 {
	return stlhash.Hash(self.path)
}

func (self *Package) GetExternPackage(name string) (*Package, bool) {
	if len(name) == 0 {
		return nil, false
	}
	pkg := stlslices.Last(self.externs.Get(name))
	return pkg, pkg != nil
}

func (self *Package) GetLinkedPackages() []*Package {
	return self.externs.Get("")
}

func (self *Package) GetIdent(name string, allowLinkedPkgs ...bool) (any, bool) {
	isAllowLinkedPkgs := stlslices.Last(allowLinkedPkgs)
	ident := self.idents.Get(name)
	if ident != nil {
		return ident, true
	}
	if isAllowLinkedPkgs {
		for _, pkg := range self.GetLinkedPackages() {
			ident, ok := pkg.GetIdent(name, false)
			if ok {
				return ident, true
			}
		}
	}
	return nil, false
}

func (self *Package) AppendGlobal(g Global) Global {
	self.globals.PushBack(g)
	switch g.(type) {
	case Function, *TypeDef, *TypeAliasDef:

	}
	return g
}
