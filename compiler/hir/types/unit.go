package types

import (
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Unit = &_UnitType{}

type UnitType interface {
	Type
	unit()
}

type _UnitType struct{}

func (*_UnitType) unit() {}

func (t *_UnitType) Print(p *hir.Printer) {
	p.WriteString(t.String())
}

func (t *_UnitType) String() string {
	return "unit"
}

func (t *_UnitType) Equal(p Type) bool {
	return stlval.Is[UnitType](p)
}
