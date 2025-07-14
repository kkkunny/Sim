package global

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// VarDef 变量定义
type VarDef struct {
	pkgGlobalAttr
	values.VarDecl
	attrs []VarAttr
	value hir.Value
}

func NewVarDef(mut bool, name string, t hir.Type, attrs ...VarAttr) *VarDef {
	return &VarDef{
		VarDecl: values.NewVarDecl(mut, name, t),
		attrs:   attrs,
	}
}

func (self *VarDef) Attrs() []VarAttr {
	return self.attrs
}

func (self *VarDef) SetValue(value hir.Value) {
	self.value = value
}

func (self *VarDef) Value() (hir.Value, bool) {
	return self.value, self.value != nil
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
