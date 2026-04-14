package ast

import "fmt"

type Type interface {
	typ()
	fmt.Stringer
}

type IdentType struct {
	Name string
}

func (t *IdentType) typ() {}

func (t *IdentType) String() string {
	return t.Name
}
