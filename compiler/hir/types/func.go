package types

import (
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type FuncType interface {
	hir.Type
	GetReturn() hir.Type
	GetParams() []hir.Type
}

type _FuncType struct {
	Return hir.Type
	Params []hir.Type
}

func NewFuncType(ret hir.Type, params ...hir.Type) FuncType {
	return &_FuncType{
		Return: ret,
		Params: params,
	}
}

func (t *_FuncType) Print(p *hir.Printer) {
	p.WriteString("(")
	for i, param := range t.Params {
		p.WriteBy(param)
		if i < len(t.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	p.WriteString(" -> ")
	p.WriteBy(t.Return)
}

func (t *_FuncType) String() string {
	var buf strings.Builder
	hir.Print(&buf, t)
	return buf.String()
}

func (t *_FuncType) Equal(p hir.Type) bool {
	dst, ok := p.(FuncType)
	if !ok {
		return false
	}
	dstParams := dst.GetParams()
	if !t.Return.Equal(dst.GetReturn()) || len(t.Params) != len(dstParams) {
		return false
	}
	return stlslices.All(t.Params, func(i int, p hir.Type) bool {
		return p.Equal(dstParams[i])
	})
}

func (t *_FuncType) GetReturn() hir.Type {
	return t.Return
}

func (t *_FuncType) GetParams() []hir.Type {
	return t.Params
}

type _CustomFuncType struct {
	_CustomBaseType[FuncType]
}

func (t *_CustomFuncType) GetReturn() hir.Type {
	return t.GetUnderlying().(FuncType).GetReturn()
}

func (t *_CustomFuncType) GetParams() []hir.Type {
	return t.GetUnderlying().(FuncType).GetParams()
}
