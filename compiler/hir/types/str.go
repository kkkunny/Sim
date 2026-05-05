package types

import (
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Str = &_StringType{}

type StringType interface {
	Type
	string()
}

type _StringType struct{}

func (t *_StringType) string() {}

func (t *_StringType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_StringType) String() string {
	return "str"
}

func (t *_StringType) Equal(p Type) bool {
	return stlval.Is[StringType](p)
}

type _CustomStringType struct {
	_CustomBaseType[StringType]
}

func (t *_CustomStringType) string() {}
