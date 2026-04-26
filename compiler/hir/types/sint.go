package types

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir"
)

var (
	I8  = &_SintType{Bits: 8}
	I16 = &_SintType{Bits: 16}
	I32 = &_SintType{Bits: 32}
	I64 = &_SintType{Bits: 64}
)

type SintType interface {
	IntegerType
	sint()
}

type _SintType struct {
	Bits uint8
}

func (*_SintType) sint()    {}
func (*_SintType) integer() {}

func (t *_SintType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_SintType) String() string {
	return fmt.Sprintf("i%d", t.Bits)
}

func (t *_SintType) Equal(p Type) bool {
	dst, ok := p.(SintType)
	if !ok {
		return false
	}
	return t.Bits == dst.GetBits()
}

func (t *_SintType) GetBits() uint8 {
	return t.Bits
}

type _CustomSintType struct {
	_CustomBaseType[SintType]
}

func (*_CustomSintType) sint()    {}
func (*_CustomSintType) integer() {}

func (t *_CustomSintType) GetBits() uint8 {
	return t.Underlying.GetBits()
}
