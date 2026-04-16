package cir

type Builder struct {
	Globals []Global
}

func NewBuilder() *Builder {
	return &Builder{}
}
