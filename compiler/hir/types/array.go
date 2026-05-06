package types

import (
	"fmt"
	"math/big"

	"github.com/kkkunny/Sim/compiler/hir"
)

type ArrayType interface {
	hir.Type
	GetSize() *big.Int
	GetElem() hir.Type
}

type _ArrayType struct {
	Size *big.Int
	Elem hir.Type
}

func NewArrayType(size *big.Int, elem hir.Type) ArrayType {
	return &_ArrayType{
		Size: size,
		Elem: elem,
	}
}

func (t *_ArrayType) Print(p *hir.Printer) {
	p.WriteFormat(t.String())
}

func (t *_ArrayType) String() string {
	return fmt.Sprintf("[%s]%s", t.Size, t.Elem)
}

func (t *_ArrayType) Equal(p hir.Type) bool {
	dst, ok := p.(ArrayType)
	if !ok {
		return false
	}
	if t.Size.String() != dst.GetSize().String() {
		return false
	}
	return t.Elem.Equal(dst.GetElem())
}

func (t *_ArrayType) GetSize() *big.Int {
	return t.Size
}

func (t *_ArrayType) GetElem() hir.Type {
	return t.Elem
}

type _CustomArrayType struct {
	_CustomBaseType[ArrayType]
}

func (t *_CustomArrayType) GetSize() *big.Int {
	return t.GetUnderlying().(ArrayType).GetSize()
}

func (t *_CustomArrayType) GetElem() hir.Type {
	return t.GetUnderlying().(ArrayType).GetElem()
}
