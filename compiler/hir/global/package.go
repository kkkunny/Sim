package global

import (
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlhash "github.com/kkkunny/stl/hash"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/config"
)

type pkgGlobalAttr struct {
	pkg *Package
	pub bool
}

func (self pkgGlobalAttr) Package() *Package {
	return self.pkg
}

func (self pkgGlobalAttr) setPackage(pkg *Package) {
	self.pkg = pkg
}

func (self pkgGlobalAttr) Public() bool {
	return self.pub
}

func (self pkgGlobalAttr) setPublic(pub bool) {
	self.pub = pub
}

type Package struct {
	path    stlos.FilePath
	globals linkedlist.LinkedList[Global]

	externs hashmap.HashMap[string, []*Package]
	idents  hashmap.HashMap[string, any]
}

func NewPackage(path stlos.FilePath) *Package {
	return &Package{
		path:    path,
		globals: linkedlist.NewLinkedList[Global](),
		externs: hashmap.StdWith[string, []*Package](),
	}
}

func (self *Package) String() string {
	return string(self.path)
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

func (self *Package) AppendGlobal(pub bool, g Global) Global {
	g.setPackage(self)
	g.setPublic(pub)
	self.globals.PushBack(g)
	switch g.(type) {
	case Function, *CustomTypeDef, *AliasTypeDef:

	}
	return g
}

// IsIn 是否处于目标包下
func (self *Package) IsIn(dst *Package) bool {
	rel, err := self.path.Rel(dst.path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(string(rel), "..")
}

// IsBuildIn 是否是buildin包
func (self *Package) IsBuildIn() bool {
	return self.path == config.BuildInPkgPath
}

func (self *Package) Path() stlos.FilePath {
	return self.path
}
