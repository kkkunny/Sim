package global

import "github.com/kkkunny/Sim/compiler/hir/types"

// Function 函数定义
type Function interface {
	Global
	Params() []*Param
	Attrs() []FuncAttr
}

// Param 函数形参
type Param struct {
	mut  bool
	typ  types.Type
	name string
}

func (self *Param) Type() types.Type {
	return self.typ
}

func (self *Param) Mutable() bool {
	return self.mut
}

func (self *Param) Name() (string, bool) {
	return self.name, self.name != ""
}

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
func (self FuncAttrLinkName) funcAttr() {}

// FuncAttrInline 内联
type FuncAttrInline struct {
	inline bool
}

func WithInlineFuncAttr(inline bool) FuncAttr {
	return &FuncAttrInline{inline: inline}
}
func (self FuncAttrInline) funcAttr() {}

// FuncAttrVararg 可变参数
type FuncAttrVararg struct{}

func WithVarargFuncAttr() FuncAttr {
	return &FuncAttrVararg{}
}
func (self FuncAttrVararg) funcAttr() {}
