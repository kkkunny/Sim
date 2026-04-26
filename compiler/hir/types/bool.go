package types

import (
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Bool = &_BooleanType{}

type BooleanType interface {
	Type
	boolean()
}

type _BooleanType struct{}

func (t *_BooleanType) boolean() {}

func (t *_BooleanType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_BooleanType) String() string {
	return "bool"
}

func (t *_BooleanType) Equal(p Type) bool {
	return stlval.Is[BooleanType](p)
}

type _CustomBooleanType struct {
	_CustomBaseType[BooleanType]
}

func (t *_CustomBooleanType) boolean() {}
