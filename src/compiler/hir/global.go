package hir

// Global 全局
type Global interface {
	global()
}

// Function 函数
type Function struct {
	// 属性
	ExternName string // 外部名
	NoReturn   bool   // 函数是否不返回
	Inline     *bool  // 函数是否强制内联或者强制不内联
	Init, Fini bool   // 是否是init or fini函数

	Ret    Type
	Params []*Param
	VarArg bool
	Body   *Block // 可能为空
}

func (self Function) global() {}

func (self Function) stmt() {}

func (self Function) ident() {}

func (self Function) GetType() Type {
	paramTypes := make([]Type, len(self.Params))
	for i, p := range self.Params {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes, self.VarArg)
}

func (self Function) GetMut() bool {
	return false
}

func (self Function) IsTemporary() bool {
	return true
}

func (self Function) GetMethodType() *TypeFunc {
	paramTypes := make([]Type, len(self.Params)-1)
	for i, p := range self.Params[1:] {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes, self.VarArg)
}

// GlobalVariable 全局变量
type GlobalVariable struct {
	ExternName string

	Type  Type
	Value Expr // 可能为空
}

func (self GlobalVariable) global() {}

func (self GlobalVariable) stmt() {}

func (self GlobalVariable) ident() {}

func (self GlobalVariable) GetType() Type {
	return self.Type
}

func (self GlobalVariable) GetMut() bool {
	return true
}

func (self GlobalVariable) IsTemporary() bool {
	return false
}
