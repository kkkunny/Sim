package hir

type Global interface {
	printWriter
	global()
}
