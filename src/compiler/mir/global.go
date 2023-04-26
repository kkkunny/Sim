package mir

import (
	"github.com/kkkunny/containers/list"
)

// Global 全局
type Global interface {
	global()
}

// Alias 类型别名
type Alias struct {
	Name   string
	Target Type
}

func (self *Package) NewAlias(name string, t Type) *Alias {
	if t.IsVoid() {
		panic("unreachable")
	}
	g := &Alias{
		Name:   name,
		Target: t,
	}
	self.Globals.PushBack(g)
	return g
}
func (self Alias) global() {}

// Function 函数
type Function struct {
	Type   Type
	Name   string
	Params []*Param
	Blocks *list.List[*Block]

	Extern   bool
	inline   bool
	noInline bool
	NoReturn bool
	init     bool
	fini     bool
}

func (self *Package) NewFunction(t Type, name string) *Function {
	if !t.IsFunc() {
		panic("unreachable")
	}
	paramTypes := t.GetFuncParams()
	g := &Function{
		Type:   t,
		Name:   name,
		Params: make([]*Param, len(paramTypes)),
		Blocks: list.NewList[*Block](),
	}
	self.Globals.PushBack(g)
	for i, p := range paramTypes {
		g.Params[i] = g.newParam(p)
	}
	return g
}
func (self Function) global() {}
func (self Function) GetType() Type {
	return self.Type
}
func (self *Function) SetInline(v bool) {
	if self.noInline && v {
		panic("unreachable")
	}
	self.inline = v
}
func (self Function) GetInline() bool {
	return self.inline
}
func (self *Function) SetNoInline(v bool) {
	if self.inline && v {
		panic("unreachable")
	}
	self.noInline = v
}
func (self Function) GetNoInline() bool {
	return self.noInline
}
func (self *Function) SetInit(v bool) {
	if v && len(self.Type.GetFuncParams()) != 0 {
		panic("unreachable")
	}
	self.init = v
}
func (self Function) GetInit() bool {
	return self.init
}
func (self *Function) SetFini(v bool) {
	if v && len(self.Type.GetFuncParams()) != 0 {
		panic("unreachable")
	}
	self.fini = v
}
func (self Function) GetFini() bool {
	return self.fini
}

// Variable 变量
type Variable struct {
	Type  Type
	Name  string
	Value Constant // 可能为空

	Extern bool
}

func (self *Package) NewVariable(t Type, name string, value Constant) *Variable {
	if value != nil && !t.Equal(value.GetType()) {
		panic("unreachable")
	}
	g := &Variable{
		Type:  t,
		Name:  name,
		Value: value,
	}
	self.Globals.PushBack(g)
	return g
}
func (self Variable) global() {}
func (self Variable) GetType() Type {
	return NewTypePtr(self.Type)
}
func (self Variable) constant() {}
