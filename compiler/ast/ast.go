package ast

type Ast interface {
	ast()
}

type FuncDecl struct {
	Name string
}

func (f FuncDecl) ast() {}
