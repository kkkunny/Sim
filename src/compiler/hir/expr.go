package hir

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
	GetMut() bool
	IsTemporary() bool
}

// Ident 标识符
type Ident interface {
	Expr
	ident()
}

// Integer 整数
type Integer struct {
	Type  Type
	Value int64
}

func (self Integer) stmt() {}

func (self Integer) GetType() Type {
	return self.Type
}

func (self Integer) GetMut() bool {
	return false
}

func (self Integer) IsTemporary() bool {
	return true
}

// Float 浮点数
type Float struct {
	Type  Type
	Value float64
}

func (self Float) stmt() {}

func (self Float) GetType() Type {
	return self.Type
}

func (self Float) GetMut() bool {
	return false
}

func (self Float) IsTemporary() bool {
	return true
}

// Boolean 布尔数
type Boolean struct {
	Type  Type
	Value bool
}

func (self Boolean) stmt() {}

func (self Boolean) GetType() Type {
	return self.Type
}

func (self Boolean) GetMut() bool {
	return false
}

func (self Boolean) IsTemporary() bool {
	return true
}

// String 字符串
type String struct {
	Type  Type
	Value string
}

func (self String) stmt() {}

func (self String) GetType() Type {
	return self.Type
}

func (self String) GetMut() bool {
	return false
}

func (self String) IsTemporary() bool {
	return true
}

// Null 空指针
type Null struct {
	Type Type
}

func (self Null) stmt() {}

func (self Null) GetType() Type {
	return self.Type
}

func (self Null) GetMut() bool {
	return false
}

func (self Null) IsTemporary() bool {
	return true
}

// Binary 二元表达式
type Binary struct {
	Opera       string
	Left, Right Expr
}

func (self Binary) stmt() {}

func (self Binary) GetType() Type {
	return self.Left.GetType()
}

func (self Binary) GetMut() bool {
	return false
}

func (self Binary) IsTemporary() bool {
	return true
}

// FuncCall 函数调用
type FuncCall struct {
	Func Expr
	Args []Expr
}

func (self FuncCall) stmt() {}

func (self FuncCall) GetType() Type {
	return GetBaseType(self.Func.GetType()).(*TypeFunc).Ret
}

func (self FuncCall) GetMut() bool {
	return false
}

func (self FuncCall) IsTemporary() bool {
	return true
}

// MethodCall 方法调用
type MethodCall struct {
	Method *Method
	Args   []Expr
}

func (self MethodCall) stmt() {}

func (self MethodCall) GetType() Type {
	return self.Method.GetMethodType().Ret
}

func (self MethodCall) GetMut() bool {
	return false
}

func (self MethodCall) IsTemporary() bool {
	return true
}

// Param 参数
type Param struct {
	Type Type
}

func (self Param) stmt() {}

func (self Param) ident() {}

func (self Param) GetType() Type {
	return self.Type
}

func (self Param) GetMut() bool {
	return true
}

func (self Param) IsTemporary() bool {
	return false
}

// Array 数组
type Array struct {
	Type  Type
	Elems []Expr
}

func (self Array) stmt() {}

func (self Array) GetType() Type {
	return self.Type
}

func (self Array) GetMut() bool {
	return false
}

func (self Array) IsTemporary() bool {
	return true
}

// EmptyArray 空数组
type EmptyArray struct {
	Type Type
}

func (self EmptyArray) stmt() {}

func (self EmptyArray) GetType() Type {
	return self.Type
}

func (self EmptyArray) GetMut() bool {
	return false
}

func (self EmptyArray) IsTemporary() bool {
	return true
}

// Assign 赋值
type Assign struct {
	Opera       string
	Left, Right Expr
}

func (self Assign) stmt() {}

func (self Assign) GetType() Type {
	return None
}

func (self Assign) GetMut() bool {
	return false
}

func (self Assign) IsTemporary() bool {
	return true
}

// Equal 赋值
type Equal struct {
	Opera       string
	Left, Right Expr
}

func (self Equal) stmt() {}

func (self Equal) GetType() Type {
	return Bool
}

func (self Equal) GetMut() bool {
	return false
}

func (self Equal) IsTemporary() bool {
	return true
}

// Unary 一元表达式
type Unary struct {
	Mutable bool
	Type    Type
	Opera   string
	Value   Expr
}

func (self Unary) stmt() {}

func (self Unary) GetType() Type {
	return self.Type
}

func (self Unary) GetMut() bool {
	return self.Mutable
}

func (self Unary) IsTemporary() bool {
	return true
}

// Index 索引
type Index struct {
	Type        Type
	From, Index Expr
}

func (self Index) stmt() {}

func (self Index) GetType() Type {
	return self.Type
}

func (self Index) GetMut() bool {
	return self.From.GetMut()
}

func (self Index) IsTemporary() bool {
	return self.From.IsTemporary()
}

// Select 选择
type Select struct {
	Cond, True, False Expr
}

func (self Select) stmt() {}

func (self Select) GetType() Type {
	return self.True.GetType()
}

func (self Select) GetMut() bool {
	return self.True.GetMut() && self.False.GetMut()
}

func (self Select) IsTemporary() bool {
	return self.True.IsTemporary() || self.False.IsTemporary()
}

// Tuple 元组
type Tuple struct {
	Type  Type
	Elems []Expr
}

func (self Tuple) stmt() {}

func (self Tuple) GetType() Type {
	return self.Type
}

func (self Tuple) GetMut() bool {
	return false
}

func (self Tuple) IsTemporary() bool {
	return true
}

// EmptyTuple 空元组
type EmptyTuple struct {
	Type Type
}

func (self EmptyTuple) stmt() {}

func (self EmptyTuple) GetType() Type {
	return self.Type
}

func (self EmptyTuple) GetMut() bool {
	return false
}

func (self EmptyTuple) IsTemporary() bool {
	return true
}

// Struct 结构体
type Struct struct {
	Type   Type
	Fields []Expr
}

func (self Struct) stmt() {}

func (self Struct) GetType() Type {
	return self.Type
}

func (self Struct) GetMut() bool {
	return false
}

func (self Struct) IsTemporary() bool {
	return true
}

// EmptyStruct 空结构体
type EmptyStruct struct {
	Type Type
}

func (self EmptyStruct) stmt() {}

func (self EmptyStruct) GetType() Type {
	return self.Type
}

func (self EmptyStruct) GetMut() bool {
	return false
}

func (self EmptyStruct) IsTemporary() bool {
	return true
}

// GetField 获取成员
type GetField struct {
	From  Expr
	Index string
}

func (self GetField) stmt() {}

func (self GetField) GetType() Type {
	return GetBaseType(self.From.GetType()).(*TypeStruct).Fields.Get(self.Index).Second
}

func (self GetField) GetMut() bool {
	return self.From.GetMut()
}

func (self GetField) IsTemporary() bool {
	return self.From.IsTemporary()
}

// Covert 类型转换
type Covert struct {
	From Expr
	To   Type
}

func (self Covert) stmt() {}

func (self Covert) GetType() Type {
	return self.To
}

func (self Covert) GetMut() bool {
	return false
}

func (self Covert) IsTemporary() bool {
	return true
}

// Method 方法
type Method struct {
	Self Expr // 类型定义 || 类型定义指针
	Func *Function
}

func (self Method) stmt() {}

func (self Method) GetType() Type {
	return self.Func.GetType()
}

func (self Method) GetMethodType() *TypeFunc {
	return self.Func.GetMethodType()
}

func (self Method) GetMut() bool {
	return false
}

func (self Method) IsTemporary() bool {
	return true
}

// Alloc 栈内存分配
type Alloc struct {
	Size Expr
}

func (self Alloc) stmt() {}

func (self Alloc) GetType() Type {
	return NewPtrType(Usize)
}

func (self Alloc) GetMut() bool {
	return false
}

func (self Alloc) IsTemporary() bool {
	return true
}

// GetInterfaceField 获取接口成员
type GetInterfaceField struct {
	From  Expr
	Index string
}

func (self GetInterfaceField) stmt() {}

func (self GetInterfaceField) GetType() Type {
	ft := self.GetType().(*TypeFunc)
	ft.Params = append([]Type{NewPtrType(Usize)}, ft.Params...)
	return ft
}

func (self GetInterfaceField) GetMethodType() *TypeFunc {
	return GetBaseType(self.From.GetType()).(*TypeInterface).Fields.Get(self.Index)
}

func (self GetInterfaceField) GetMut() bool {
	return false
}

func (self GetInterfaceField) IsTemporary() bool {
	return true
}

// InterfaceFieldCall 接口成员调用
type InterfaceFieldCall struct {
	Field *GetInterfaceField
	Args  []Expr
}

func (self InterfaceFieldCall) stmt() {}

func (self InterfaceFieldCall) GetType() Type {
	return self.Field.GetMethodType().Ret
}

func (self InterfaceFieldCall) GetMut() bool {
	return false
}

func (self InterfaceFieldCall) IsTemporary() bool {
	return true
}
