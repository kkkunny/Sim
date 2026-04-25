package types

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir"
)

var (
	F32 = &_FloatType{Bits: 32}
	F64 = &_FloatType{Bits: 64}
)

type FloatType interface {
	NumberType
	float()
}

type _FloatType struct {
	Bits uint8
}

func (*_FloatType) float() {}

func (t *_FloatType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_FloatType) String() string {
	return fmt.Sprintf("f%d", t.Bits)
}

func (t *_FloatType) Equal(p Type) bool {
	dst, ok := p.(FloatType)
	if !ok {
		return false
	}
	return t.Bits == dst.GetBits()
}

func (t *_FloatType) GetBits() uint8 {
	return t.Bits
}
