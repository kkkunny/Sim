package local

import (
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

type VarDef interface {
	Local
	values.Ident
	SetEscaped(v bool)
	Escaped() bool
}

// VarDecl 变量声明
type VarDecl struct {
	values.VarDecl
	escaped bool
}

func NewVarDecl(mut bool, name string, t types.Type) *VarDecl {
	return &VarDecl{VarDecl: *values.NewVarDecl(mut, name, t)}
}

func (self *VarDecl) local() {}

func (self *VarDecl) SetEscaped(v bool) {
	self.escaped = v
}

func (self *VarDecl) Escaped() bool {
	return self.escaped
}

func (self *VarDecl) Ident() {}

// Param 函数形参
type Param struct {
	VarDecl
}

func NewParam(mut bool, name string, t types.Type) *Param {
	return &Param{
		VarDecl: *NewVarDecl(mut, name, t),
	}
}
