package hir

type Global interface {
	File() *File
	Package() *Package
	SetFile(file *File)
	Public() bool
	SetPublic(pub bool)
}
