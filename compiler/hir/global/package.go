package global

import "github.com/kkkunny/Sim/compiler/hir"

type pkgGlobalAttr struct {
	pkg *hir.Package
	pub bool
}

func (self pkgGlobalAttr) Package() *hir.Package {
	return self.pkg
}

func (self *pkgGlobalAttr) SetPackage(pkg *hir.Package) {
	self.pkg = pkg
}

func (self pkgGlobalAttr) Public() bool {
	return self.pub
}

func (self *pkgGlobalAttr) SetPublic(pub bool) {
	self.pub = pub
}
