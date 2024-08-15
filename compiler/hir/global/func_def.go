package global

import (
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// FuncDef 函数定义
type FuncDef struct {
	pkgGlobalAttr
	funcType *types.FuncType
	name     string
	attrs    []FuncAttr
	params   []*Param
	body     *local.Block
}

func (self *FuncDef) FuncType() *types.FuncType {
	return self.funcType
}

func (self *FuncDef) Name() string {
	return self.name
}

func (self *FuncDef) Attrs() []FuncAttr {
	return self.attrs
}

func (self *FuncDef) Params() []*Param {
	return self.params
}

func (self *FuncDef) Block() (*local.Block, bool) {
	return self.body, self.body != nil
}

func (self *FuncDef) Type() types.Type {
	return self.FuncType()
}

func (self *FuncDef) Mutable() bool {
	return false
}
