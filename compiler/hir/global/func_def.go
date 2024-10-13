package global

import (
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// FuncDef 函数定义
type FuncDef struct {
	pkgGlobalAttr
	typ    types.FuncType
	name   string
	attrs  []FuncAttr
	params []*local.Param
	body   *local.Block
}

func (self *FuncDef) CallableType() types.CallableType {
	return self.typ
}

func (self *FuncDef) Name() string {
	return self.name
}

func (self *FuncDef) Attrs() []FuncAttr {
	return self.attrs
}

func (self *FuncDef) Params() []*local.Param {
	return self.params
}

func (self *FuncDef) Block() (*local.Block, bool) {
	return self.body, self.body != nil
}

func (self *FuncDef) Type() types.Type {
	return self.CallableType()
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

func (self *FuncDef) GetName() (string, bool) {
	return self.name, true
}
