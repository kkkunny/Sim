package global

type Package struct {
	module *Module
}

func (self *Package) Module() *Module {
	return self.module
}

func (self *Package) String() string {
	// TODO
	panic("unreachable")
}

func (self *Package) Equal(dst *Package) bool {
	// TODO
	panic("unreachable")
}

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
