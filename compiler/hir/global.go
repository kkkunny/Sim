package hir

type Global interface {
	Package() *Package
	SetPackage(pkg *Package)
	Public() bool
	SetPublic(pub bool)
}
