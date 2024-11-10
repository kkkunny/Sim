package global

import (
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// VarDef 变量定义
type VarDef struct {
	pkgGlobalAttr
	*values.VarDecl
	attrs []VarAttr
	value values.Value
}

func NewVarDef(decl *values.VarDecl, attrs ...VarAttr) *VarDef {
	return &VarDef{
		VarDecl: decl,
		attrs:   attrs,
	}
}

func (self *VarDef) Attrs() []VarAttr {
	return self.attrs
}

func (self *VarDef) SetValue(value values.Value) {
	self.value = value
}

func (self *VarDef) Value() (values.Value, bool) {
	return self.value, self.value != nil
}

func (self *VarDef) Name() string {
	return stlval.IgnoreWith(self.VarDecl.GetName())
}

// VarAttr 变量属性
type VarAttr interface {
	varAttr()
}

// VarAttrLinkName 链接名
type VarAttrLinkName struct {
	name string
}

func WithLinkNameVarAttr(name string) VarAttr {
	return &VarAttrLinkName{name: name}
}
func (self *VarAttrLinkName) varAttr() {}
func (self *VarAttrLinkName) Name() string {
	return self.name
}
