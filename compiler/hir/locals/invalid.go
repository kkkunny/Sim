package locals

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Invalid 分析失败时的占位表达式：仅用于错误恢复，存在错误时不会进入代码生成。
type Invalid struct{}

func NewInvalid() *Invalid {
	return &Invalid{}
}

func (*Invalid) expr()  {}
func (*Invalid) local() {}

func (e *Invalid) Print(p *hir.Printer) {
	p.WriteString("<invalid>")
}

func (e *Invalid) GetType() hir.Type {
	return types.Invalid
}

func (e *Invalid) Mutable() bool {
	return false
}

func (e *Invalid) Temporary() bool {
	return true
}
