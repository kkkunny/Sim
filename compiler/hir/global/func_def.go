package global

import (
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// FuncDef 函数定义
type FuncDef struct {
	pkgGlobalAttr
	*FuncDecl
	genericParams []types.GenericParamType
	attrs         []FuncAttr
	body          *local.Block
}

func NewFuncDef(decl *FuncDecl, genericParams []types.GenericParamType, attrs ...FuncAttr) *FuncDef {
	return &FuncDef{
		FuncDecl:      decl,
		genericParams: genericParams,
		attrs:         attrs,
	}
}

func (self *FuncDef) GenericParams() []types.GenericParamType {
	return self.genericParams
}

func (self *FuncDef) Attrs() []FuncAttr {
	return self.attrs
}

func (self *FuncDef) Body() (*local.Block, bool) {
	return self.body, self.body != nil
}

func (self *FuncDef) Mutable() bool {
	return false
}

func (self *FuncDef) SetBody(b *local.Block) {
	self.body = b
}

func (self *FuncDef) Storable() bool {
	return false
}

func (self *FuncDef) Parent() local.Scope {
	return self.pkg
}
