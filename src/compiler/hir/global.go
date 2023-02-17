package hir

// Global 全局
type Global interface {
	global()
}

// Typedef 类型定义
type Typedef struct {
	Pkg    PkgPath // 包
	Name   string  // 类型名
	Target Type    // 目标类型

	Methods map[string]*Method // 方法
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
	Typ   Type   // 类型
	Name  string // 名
	Value Expr   // 值
}

func NewGlobalValue(t Type, name string, v Expr) *GlobalValue {
	return &GlobalValue{
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

func (self GlobalValue) Immediate() bool {
	return false
}

func (self GlobalValue) ident() {}

// Function 函数
type Function struct {
	Typ  Type   // 类型
	Name string // 名
	Body *Block // 函数体（可能为空）

	// 属性
	NoReturn     bool // 函数不返回
	MustInline   bool // 强制内联
	MustNoInline bool // 强制不内联
	Init         bool // 是否是init
	Fini         bool // 是否是fini
}

func NewFunction(t Type, name string, b *Block) *Function {
	return &Function{
		Typ:  t,
		Name: name,
		Body: b,
	}
}

func (self Function) global() {}

func (self Function) stmt() {}

func (self Function) Type() Type {
	return self.Typ
}

func (self Function) Immediate() bool {
	return true
}

func (self Function) ident() {}

// Method 方法
type Method struct {
	Typ  Type     // 类型
	Self *Typedef // self
	Body *Block   // 函数体

	// 属性
	NoReturn     bool // 函数不返回
	MustInline   bool // 强制内联
	MustNoInline bool // 强制不内联
}

func NewMethod(t Type, self *Typedef, b *Block) *Method {
	return &Method{
		Typ:  t,
		Self: self,
		Body: b,
	}
}

func (self Method) global() {}

func (self Method) stmt() {}

func (self Method) Type() Type {
	return self.Typ
}

func (self Method) Immediate() bool {
	return true
}

func (self Method) ident() {}

// FunctionType 获取真实函数类型
func (self Method) FunctionType() Type {
	return NewTypeFunc(
		self.Typ.GetFuncRet(),
		append([]Type{NewTypePtr(NewTypeTypedef(self.Self))}, self.Typ.GetFuncParams()...)...,
	)
}
