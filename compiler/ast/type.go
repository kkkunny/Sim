package ast

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/token"
)

type Type interface {
	typ()
	fmt.Stringer
}

type IdentType struct {
	Name token.Token
}

func (t *IdentType) typ() {}

func (t *IdentType) String() string {
	return t.Name.OriginText
}
