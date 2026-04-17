package hir

type Scope struct {
	Parent *Scope
	Types  map[string]Type
}
