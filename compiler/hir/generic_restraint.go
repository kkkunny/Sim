package hir

// GenericRestraint 泛型约束
type GenericRestraint interface {
	GetMethodType(name string) (Type, bool)
}
