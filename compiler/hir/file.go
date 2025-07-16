package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/utils"
	"github.com/kkkunny/Sim/compiler/util"
)

type File struct {
	path stlos.FilePath
	pkg  *Package

	externs hashmap.HashMap[string, []*Package]
}

func NewFile(path stlos.FilePath, pkg *Package) *File {
	return &File{
		path:    path,
		pkg:     pkg,
		externs: hashmap.StdWith[string, []*Package](),
	}
}

func (self *File) String() string {
	return string(self.path)
}

func (self *File) Equal(dst *File) bool {
	return self.path == dst.path
}

func (self *File) Hash() uint64 {
	return util.StringHashFunc(string(self.path))
}

func (self *File) Package() *Package {
	return self.pkg
}

func (self *File) SetExternPackage(name string, pkg *Package) {
	self.externs.Set(name, []*Package{pkg})
}

func (self *File) GetExternPackage(name string) (*Package, bool) {
	if len(name) == 0 {
		return nil, false
	}
	pkg := stlslices.Last(self.externs.Get(name))
	return pkg, pkg != nil
}

func (self *File) AddLinkedPackage(pkg *Package) {
	self.externs.Set("", append(self.GetLinkedPackages(), pkg))
}

func (self *File) GetLinkedPackages() []*Package {
	return self.externs.Get("")
}

func (self *File) SetIdent(name string, ident any) bool {
	if self.externs.Contain(name) {
		return false
	}
	return self.pkg.SetIdent(name, ident)
}

func (self *File) GetIdent(name string, allowLinkedPkgs ...bool) (any, bool) {
	ident, ok := self.pkg.GetIdent(name, allowLinkedPkgs...)
	if ok {
		return ident, true
	}
	if stlslices.Last(allowLinkedPkgs) {
		for _, pkg := range self.GetLinkedPackages() {
			if ident, ok = pkg.GetIdent(name, false); ok {
				return ident, true
			}
		}
	}
	return nil, false
}

func (self *File) Path() stlos.FilePath {
	return self.path
}

func (self *File) AppendGlobal(pub bool, g Global) Global {
	g.SetFile(self)
	g.SetPublic(pub)
	self.Package().globals.PushBack(g)
	if namedGlobal, ok := g.(utils.Named); ok && !stlval.Is[utils.NotGlobalNamed](namedGlobal) {
		if name, ok := namedGlobal.GetName(); ok {
			self.Package().idents.Set(name.Value, g)
		}
	}
	return g
}
