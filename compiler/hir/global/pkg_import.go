package global

// PkgImport 包导入
type PkgImport struct {
	pkgGlobalAttr
	importAll bool
	alias     string
	pkg       *Package
}

func (self *PkgImport) ImportAll() bool {
	return self.importAll
}

func (self *PkgImport) Alias() (string, bool) {
	return self.alias, self.alias != ""
}

func (self *PkgImport) TargetPackage() *Package {
	return self.pkg
}
