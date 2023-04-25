package hir

import "github.com/kkkunny/stl/table"

// Global 全局
type Global interface {
	global()
}

// Typedef 类型定义
type Typedef struct {
	Pkg    PkgPath // 包
	Name   string  // 类型名
	Target Type    // 目标类型

	Methods     map[string]*Method                 // 方法
	GenericArgs *table.LinkedHashMap[string, Type] // 泛型参数
}

func NewTypedef(pkg PkgPath, name string, t Type) *Typedef {
	return &Typedef{
		Pkg:    pkg,
		Name:   name,
		Target: t,

		Methods: make(map[string]*Method),
	}
}

func (self Typedef) global() {}

func (self *Typedef) DeclMethod(name string, method *Method) bool {
	if _, ok := self.Methods[name]; ok {
		return false
	}
	self.Methods[name] = method
	return true
}

func (self *Typedef) LookupMethod(name string) (*Method, bool) {
	method, ok := self.Methods[name]
	return method, ok
}

// GlobalValue 全局变量
type GlobalValue struct {
	Mut   bool   // 是否可变
	Typ   Type   // 类型
	Name  string // 名
	Value Expr   // 值（可能为空）
}

func NewGlobalValue(mut bool, t Type, name string, v Expr) *GlobalValue {
	return &GlobalValue{
		Mut:   mut,
		Typ:   t,
		Name:  name,
		Value: v,
	}
}

func (self GlobalValue) global() {}

func (self GlobalValue) stmt() {}

func (self GlobalValue) Type() Type {
	return self.Typ
}

func (self GlobalValue) Mutable() bool {
	return self.Mut
}

func (self GlobalValue) Immediate() bool {
	return false
}

func (self GlobalValue) ident() {}

// Function 函数
type Function struct {
	Typ    Type     // 类型
	Name   string   // 名
	Params []*Param // 参数
	Body   *Block   // 函数体（可能为空）

	// 属性
	NoReturn     bool // 函数不返回
	MustInline   bool // 强制内联
	MustNoInline bool // 强制不内联
	Init         bool // 是否是init
	Fini         bool // 是否是fini
}

func NewFunction(t Type, name string, params []*Param, b *Block) *Function {
	return &Function{
		Typ:    t,
		Name:   name,
		Params: params,
		Body:   b,
	}
}

func (self Function) global() {}

func (self Function) stmt() {}

func (self Function) Type() Type {
	return self.Typ
}

func (self Function) Mutable() bool {
	return false
}

func (self Function) Immediate() bool {
	return true
}

func (self Function) ident() {}

// Method 方法
type Method struct {
	Typ     Type     // 类型
	SelfMut bool     // self可变性
	Self    *Typedef // self
	Params  []*Param // 参数
	Body    *Block   // 函数体

	// 属性
	NoReturn     bool // 函数不返回
	MustInline   bool // 强制内联
	MustNoInline bool // 强制不内联
}

func NewMethod(t Type, mut bool, self *Typedef, params []*Param, b *Block) *Method {
	return &Method{
		Typ:     t,
		SelfMut: mut,
		Self:    self,
		Params:  params,
		Body:    b,
	}
}

func (self Method) global() {}

func (self Method) stmt() {}

func (self Method) Type() Type {
	return self.Typ
}

func (self Method) Mutable() bool {
	// 表达式可变性，与上面的self可变性含义不同
	return false
}

func (self Method) Immediate() bool {
	return true
}

func (self Method) ident() {}

// FunctionType 获取真实函数类型
func (self Method) FunctionType() Type {
	return NewTypeFunc(
		self.Typ.GetFuncVarArg(),
		self.Typ.GetFuncRet(),
		append([]Type{NewTypePtr(NewTypeTypedef(self.Self))}, self.Typ.GetFuncParams()...)...,
	)
}
