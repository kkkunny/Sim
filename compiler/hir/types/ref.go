package types

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir"
)

type RefType interface {
	Type
	Mutable() bool
	PtrTo() Type
}

type _RefType struct {
	Mut  bool
	Elem Type
}

func NewRefType(mut bool, elem Type) RefType {
	return &_RefType{Mut: mut, Elem: elem}
}

func (t *_RefType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_RefType) String() string {
	if t.Mut {
		return fmt.Sprintf("&mut %s", t.Elem)
	}
	return fmt.Sprintf("&%s", t.Elem)
}

func (t *_RefType) Equal(p Type) bool {
	dst, ok := p.(RefType)
	if !ok {
		return false
	}
	if t.Mut != dst.Mutable() {
		return false
	}
	return t.Elem.Equal(dst.PtrTo())
}

func (t *_RefType) Mutable() bool {
	return t.Mut
}

func (t *_RefType) PtrTo() Type {
	return t.Elem
}
