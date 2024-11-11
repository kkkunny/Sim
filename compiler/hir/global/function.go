package global

import (
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// Function 函数定义
type Function interface {
	Global
	local.CallableDef
	values.Ident
	SetBody(b *local.Block)
	Attrs() []FuncAttr
}

// FuncDecl 函数声明
type FuncDecl struct {
	typ    types.FuncType
	name   string
	params []*local.Param
}

func NewFuncDecl(typ types.FuncType, name string, params ...*local.Param) *FuncDecl {
	return &FuncDecl{typ: typ, name: name, params: params}
}

func (self *FuncDecl) CallableType() types.CallableType {
	return self.typ
}

func (self *FuncDecl) Name() string {
	return self.name
}

func (self *FuncDecl) Params() []*local.Param {
	return self.params
}

func (self *FuncDecl) Type() types.Type {
	return self.CallableType()
}

func (self *FuncDecl) Ident() {}

// FuncAttr 函数属性
type FuncAttr interface {
	funcAttr()
}

// FuncAttrLinkName 链接名
type FuncAttrLinkName struct {
	name string
}

func WithLinkNameFuncAttr(name string) FuncAttr {
	return &FuncAttrLinkName{name: name}
}
func (self *FuncAttrLinkName) funcAttr() {}
func (self *FuncAttrLinkName) Name() string {
	return self.name
}

// FuncAttrInline 内联
type FuncAttrInline struct {
	inline bool
}

func WithInlineFuncAttr(inline bool) FuncAttr {
	return &FuncAttrInline{inline: inline}
}
func (self *FuncAttrInline) funcAttr() {}
func (self *FuncAttrInline) Inline() bool {
	return self.inline
}

// FuncAttrVararg 可变参数
type FuncAttrVararg struct{}

func WithVarargFuncAttr() FuncAttr {
	return &FuncAttrVararg{}
}
func (self *FuncAttrVararg) funcAttr() {}
