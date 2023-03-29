package mir

import (
	"fmt"
	"strings"

	"github.com/bahlo/generic-list-go"
)

// Global 全局
type Global interface {
	fmt.Stringer
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
func (self Alias) String() string {
	return fmt.Sprintf("type %s = %s", self.GetName(), self.Target)
}
func (self Alias) global() {}
func (self *Alias) GetName() string {
	if self.Name != "" {
		return self.Name
	}
	return fmt.Sprintf("t%p", self)
}

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
		Blocks: list.New[*Block](),
	}
	self.Globals.PushBack(g)
	for i, p := range paramTypes {
		g.Params[i] = g.newParam(uint(i), p)
	}
	return g
}
func (self Function) String() string {
	var buf strings.Builder
	buf.WriteString("func ")
	buf.WriteString(self.GetName())
	buf.WriteByte('(')
	for i, p := range self.Params {
		buf.WriteString(p.GetName())
		buf.WriteByte(' ')
		buf.WriteString(p.Type.String())
		if i < len(self.Params)-1 {
			buf.WriteString(", ")
		}
	}
	if self.Type.GetFuncVarArg() {
		buf.WriteString(", ...")
	}
	buf.WriteByte(')')
	buf.WriteString(self.Type.GetFuncRet().String())
	if self.Blocks.Len() != 0 {
		buf.WriteByte('{')
	}
	if self.Extern {
		buf.WriteString(" #extern")
	}
	if self.inline {
		buf.WriteString(" #inline")
	}
	if self.noInline {
		buf.WriteString(" #noinline")
	}
	if self.NoReturn {
		buf.WriteString(" #noreturn")
	}
	if self.init {
		buf.WriteString(" #init")
	}
	if self.fini {
		buf.WriteString(" #fini")
	}
	if self.Blocks.Len() != 0 {
		buf.WriteByte('\n')
		for cursor := self.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
			buf.WriteString(cursor.Value.String())
			buf.WriteByte('\n')
		}
		buf.WriteString("}")
	}
	return buf.String()
}
func (self Function) global() {}
func (self *Function) GetName() string {
	if self.Name != "" {
		return self.Name
	}
	return fmt.Sprintf("f%p", self)
}
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
func (self Variable) String() string {
	var buf strings.Builder
	buf.WriteString("var ")
	buf.WriteString(self.GetName())
	buf.WriteByte(' ')
	buf.WriteString(self.Type.String())
	if self.Value != nil {
		buf.WriteString(" = ")
		buf.WriteString(self.Value.GetName())
	}
	if self.Extern {
		buf.WriteString(" #extern")
	}
	return buf.String()
}
func (self Variable) global() {}
func (self *Variable) GetName() string {
	if self.Name != "" {
		return self.Name
	}
	return fmt.Sprintf("g%p", self)
}
func (self Variable) GetType() Type {
	return NewTypePtr(self.Type)
}
func (self Variable) constant() {}
