package hir

import (
	"path/filepath"
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/set"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/util"
)

type Package struct {
	path    stlos.FilePath
	globals linkedlist.LinkedList[Global]
	files   set.Set[*File]

	idents hashmap.HashMap[string, any]
}

func NewPackage(path stlos.FilePath) *Package {
	return &Package{
		path:    path,
		globals: linkedlist.NewLinkedList[Global](),
		files:   set.AnyHashSetWith[*File](),
		idents:  hashmap.StdWith[string, any](),
	}
}

func (self *Package) String() string {
	relpath, _ := filepath.Rel(string(config.OfficialPkgPath), string(self.path))
	if !strings.HasPrefix(relpath, "std") {
		return "main"
	}
	return strings.Join(strings.Split(relpath, string(filepath.Separator)), "::")
}

func (self *Package) Equal(dst *Package) bool {
	return self.path == dst.path
}

func (self *Package) Hash() uint64 {
	return util.StringHashFunc(string(self.path))
}

func (self *Package) GetDependencyPackages() []*Package {
	pkgs := set.AnyHashSetWith[*Package]()
	for fileIter := self.files.Iterator(); fileIter.Next(); {
		for pkgIter := fileIter.Value().externs.Iterator(); pkgIter.Next(); {
			for _, pkg := range pkgIter.Value().E2() {
				pkgs.Add(pkg)
			}
		}
	}
	return pkgs.ToSlice()
}

func (self *Package) SetIdent(name string, ident any) bool {
	if self.idents.Contain(name) {
		return false
	}
	self.idents.Set(name, ident)
	return true
}

func (self *Package) GetIdent(name string, _ ...bool) (any, bool) {
	ident := self.idents.Get(name)
	return ident, ident != nil
}

func (self *Package) AddFile(file *File) {
	self.files.Add(file)
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

func (self *Package) Globals() linkedlist.LinkedList[Global] {
	return self.globals
}

func (self *Package) Package() *Package {
	return self
}
