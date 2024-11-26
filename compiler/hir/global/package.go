package global

import "github.com/kkkunny/Sim/compiler/hir"

type pkgGlobalAttr struct {
	file *hir.File
	pub  bool
}

func (self pkgGlobalAttr) File() *hir.File {
	return self.file
}

func (self pkgGlobalAttr) Package() *hir.Package {
	return self.file.Package()
}

func (self *pkgGlobalAttr) SetFile(file *hir.File) {
	self.file = file
}

func (self pkgGlobalAttr) Public() bool {
	return self.pub
}

func (self *pkgGlobalAttr) SetPublic(pub bool) {
	self.pub = pub
}
