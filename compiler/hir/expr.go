package hir

import (
	"github.com/kkkunny/stl/container/optional"
	stlval "github.com/kkkunny/stl/value"
)

// Expr 表达式
type Expr interface {
	GetType() Type
	Mutable() bool
	Temporary() bool
}

// Ident 标识符
type Ident interface {
	Expr
	GetName() string
}

// Variable 变量
type Variable interface {
	Ident
	variable()
}

// VarDecl 变量声明
type VarDecl struct {
	Mut  bool
	Type Type
	Name string
}

func (self *VarDecl) GetName() string {
	return self.Name
}

func (self *VarDecl) GetType() Type {
	return self.Type
}

func (self *VarDecl) Mutable() bool {
	return self.Mut
}

func (*VarDecl) variable() {}

func (*VarDecl) Temporary() bool {
	return false
}

// Param 参数
type Param struct {
	Mut  bool
	Type Type
	Name optional.Optional[string]
}

func (self *Param) GetName() string {
	return stlval.TernaryAction(self.Name.IsNone(), func() string {
		return ""
	}, func() string {
		return self.Name.MustValue()
	})
}

func (self *Param) GetType() Type {
	return self.Type
}

func (self *Param) Mutable() bool {
	return self.Mut
}

func (*Param) variable() {}

func (*Param) Temporary() bool {
	return false
}
