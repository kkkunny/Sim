package types

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir"
)

var (
	U8  = &_UintType{Bits: 8}
	U16 = &_UintType{Bits: 16}
	U32 = &_UintType{Bits: 32}
	U64 = &_UintType{Bits: 64}
)

type UintType interface {
	IntegerType
	uint()
}

type _UintType struct {
	Bits uint8
}

func (*_UintType) uint()    {}
func (*_UintType) integer() {}

func (t *_UintType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_UintType) String() string {
	return fmt.Sprintf("u%d", t.Bits)
}

func (t *_UintType) Equal(p Type) bool {
	dst, ok := p.(UintType)
	if !ok {
		return false
	}
	return t.Bits == dst.GetBits()
}

func (t *_UintType) GetBits() uint8 {
	return t.Bits
}
