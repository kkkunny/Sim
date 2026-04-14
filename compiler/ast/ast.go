package ast

import "fmt"

type Ast interface {
	Ast()
	String() string
}

type FuncDecl struct {
	Name string
}

func (f *FuncDecl) Ast() {}

func (f *FuncDecl) String() string {
	return fmt.Sprintf("FuncDecl(Name: %s)", f.Name)
}
