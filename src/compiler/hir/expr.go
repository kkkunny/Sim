package hir

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
	GetMut() bool
	IsTemporary() bool
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

// Ident 标识符
type Ident interface {
	Expr
	ident()
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

// Unary 一元表达式
type Unary interface {
	Expr
	unary()
}

// Not 逻辑相反
type Not struct {
	Value Expr
}

func (self Not) stmt() {}

func (self Not) GetType() Type {
	return Bool
}

func (self Not) GetMut() bool {
	return false
}

func (self Not) IsTemporary() bool {
	return true
}

func (self Not) unary() {}

// GetPointer 取指针
type GetPointer struct {
	Value Expr
}

func (self GetPointer) stmt() {}

func (self GetPointer) GetType() Type {
	return NewPtrType(self.Value.GetType())
}

func (self GetPointer) GetMut() bool {
	return false
}

func (self GetPointer) IsTemporary() bool {
	return true
}

func (self GetPointer) unary() {}

// GetValue 取值
type GetValue struct {
	Value Expr
}

func (self GetValue) stmt() {}

func (self GetValue) GetType() Type {
	return GetBaseType(self.Value.GetType()).(*TypePtr).Elem
}

func (self GetValue) GetMut() bool {
	return self.Value.GetMut()
}

func (self GetValue) IsTemporary() bool {
	return true
}

func (self GetValue) unary() {}

// Binary 二元运算
type Binary interface {
	Expr
	binary()
}

// Assign 赋值
type Assign struct {
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

func (self Assign) binary() {}

// LogicAnd 逻辑与
type LogicAnd struct {
	Left, Right Expr
}

func (self LogicAnd) stmt() {}

func (self LogicAnd) GetType() Type {
	return self.Left.GetType()
}

func (self LogicAnd) GetMut() bool {
	return false
}

func (self LogicAnd) IsTemporary() bool {
	return true
}

func (self LogicAnd) binary() {}

// LogicOr 逻辑或
type LogicOr struct {
	Left, Right Expr
}

func (self LogicOr) stmt() {}

func (self LogicOr) GetType() Type {
	return self.Left.GetType()
}

func (self LogicOr) GetMut() bool {
	return false
}

func (self LogicOr) IsTemporary() bool {
	return true
}

func (self LogicOr) binary() {}

// Equal 相等
type Equal struct {
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

func (self Equal) binary() {}

// NotEqual 不等
type NotEqual struct {
	Left, Right Expr
}

func (self NotEqual) stmt() {}

func (self NotEqual) GetType() Type {
	return Bool
}

func (self NotEqual) GetMut() bool {
	return false
}

func (self NotEqual) IsTemporary() bool {
	return true
}

func (self NotEqual) binary() {}

// LessThan 小于
type LessThan struct {
	Left, Right Expr
}

func (self LessThan) stmt() {}

func (self LessThan) GetType() Type {
	return Bool
}

func (self LessThan) GetMut() bool {
	return false
}

func (self LessThan) IsTemporary() bool {
	return true
}

func (self LessThan) binary() {}

// LessOrEqualThan 小于等于
type LessOrEqualThan struct {
	Left, Right Expr
}

func (self LessOrEqualThan) stmt() {}

func (self LessOrEqualThan) GetType() Type {
	return Bool
}

func (self LessOrEqualThan) GetMut() bool {
	return false
}

func (self LessOrEqualThan) IsTemporary() bool {
	return true
}

func (self LessOrEqualThan) binary() {}

// Add 加
type Add struct {
	Left, Right Expr
}

func (self Add) stmt() {}

func (self Add) GetType() Type {
	return self.Left.GetType()
}

func (self Add) GetMut() bool {
	return false
}

func (self Add) IsTemporary() bool {
	return true
}

func (self Add) binary() {}

// Subtract 减
type Subtract struct {
	Left, Right Expr
}

func (self Subtract) stmt() {}

func (self Subtract) GetType() Type {
	return self.Left.GetType()
}

func (self Subtract) GetMut() bool {
	return false
}

func (self Subtract) IsTemporary() bool {
	return true
}

func (self Subtract) binary() {}

// Multiply 乘
type Multiply struct {
	Left, Right Expr
}

func (self Multiply) stmt() {}

func (self Multiply) GetType() Type {
	return self.Left.GetType()
}

func (self Multiply) GetMut() bool {
	return false
}

func (self Multiply) IsTemporary() bool {
	return true
}

func (self Multiply) binary() {}

// Divide 除
type Divide struct {
	Left, Right Expr
}

func (self Divide) stmt() {}

func (self Divide) GetType() Type {
	return self.Left.GetType()
}

func (self Divide) GetMut() bool {
	return false
}

func (self Divide) IsTemporary() bool {
	return true
}

func (self Divide) binary() {}

// Mod 取余
type Mod struct {
	Left, Right Expr
}

func (self Mod) stmt() {}

func (self Mod) GetType() Type {
	return self.Left.GetType()
}

func (self Mod) GetMut() bool {
	return false
}

func (self Mod) IsTemporary() bool {
	return true
}

func (self Mod) binary() {}

// And 与
type And struct {
	Left, Right Expr
}

func (self And) stmt() {}

func (self And) GetType() Type {
	return self.Left.GetType()
}

func (self And) GetMut() bool {
	return false
}

func (self And) IsTemporary() bool {
	return true
}

func (self And) binary() {}

// Or 或
type Or struct {
	Left, Right Expr
}

func (self Or) stmt() {}

func (self Or) GetType() Type {
	return self.Left.GetType()
}

func (self Or) GetMut() bool {
	return false
}

func (self Or) IsTemporary() bool {
	return true
}

func (self Or) binary() {}

// Xor 异或
type Xor struct {
	Left, Right Expr
}

func (self Xor) stmt() {}

func (self Xor) GetType() Type {
	return self.Left.GetType()
}

func (self Xor) GetMut() bool {
	return false
}

func (self Xor) IsTemporary() bool {
	return true
}

func (self Xor) binary() {}

// ShiftLeft 左移
type ShiftLeft struct {
	Left, Right Expr
}

func (self ShiftLeft) stmt() {}

func (self ShiftLeft) GetType() Type {
	return self.Left.GetType()
}

func (self ShiftLeft) GetMut() bool {
	return false
}

func (self ShiftLeft) IsTemporary() bool {
	return true
}

func (self ShiftLeft) binary() {}

// ShiftRight 右移
type ShiftRight struct {
	Left, Right Expr
}

func (self ShiftRight) stmt() {}

func (self ShiftRight) GetType() Type {
	return self.Left.GetType()
}

func (self ShiftRight) GetMut() bool {
	return false
}

func (self ShiftRight) IsTemporary() bool {
	return true
}

func (self ShiftRight) binary() {}

// Call 调用
type Call interface {
	Expr
	call()
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

func (self FuncCall) call() {}

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

func (self MethodCall) call() {}

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

func (self InterfaceFieldCall) call() {}

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

// Index 索引
type Index interface {
	Expr
	index()
}

// ArrayIndex 数组索引
type ArrayIndex struct {
	From, Index Expr
}

func (self ArrayIndex) stmt() {}

func (self ArrayIndex) GetType() Type {
	return GetBaseType(self.From.GetType()).(*TypeArray).Elem
}

func (self ArrayIndex) GetMut() bool {
	return self.From.GetMut()
}

func (self ArrayIndex) IsTemporary() bool {
	return self.From.IsTemporary()
}

func (self ArrayIndex) index() {}

// PointerIndex 指针索引
type PointerIndex struct {
	From, Index Expr
}

func (self PointerIndex) stmt() {}

func (self PointerIndex) GetType() Type {
	return GetBaseType(self.From.GetType()).(*TypePtr).Elem
}

func (self PointerIndex) GetMut() bool {
	return self.From.GetMut()
}

func (self PointerIndex) IsTemporary() bool {
	return self.From.IsTemporary()
}

func (self PointerIndex) index() {}

// TupleIndex 元组索引
type TupleIndex struct {
	From  Expr
	Index uint
}

func (self TupleIndex) stmt() {}

func (self TupleIndex) GetType() Type {
	return GetBaseType(self.From.GetType()).(*TypeTuple).Elems[self.Index]
}

func (self TupleIndex) GetMut() bool {
	return self.From.GetMut()
}

func (self TupleIndex) IsTemporary() bool {
	return self.From.IsTemporary()
}

func (self TupleIndex) index() {}

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
type Covert interface {
	Expr
	covert()
}

// WrapCovert 包装转换
type WrapCovert struct {
	From Expr
	To   Type
}

func (self WrapCovert) stmt() {}

func (self WrapCovert) GetType() Type {
	return self.To
}

func (self WrapCovert) GetMut() bool {
	return false
}

func (self WrapCovert) IsTemporary() bool {
	return true
}

func (self WrapCovert) covert() {}

// NumberCovert 数字转换
type NumberCovert struct {
	From Expr
	To   Type
}

func (self NumberCovert) stmt() {}

func (self NumberCovert) GetType() Type {
	return self.To
}

func (self NumberCovert) GetMut() bool {
	return false
}

func (self NumberCovert) IsTemporary() bool {
	return true
}

func (self NumberCovert) covert() {}

// Ptr2UsizeCovert 指针转换成usize
type Ptr2UsizeCovert struct {
	From Expr
	To   Type
}

func (self Ptr2UsizeCovert) stmt() {}

func (self Ptr2UsizeCovert) GetType() Type {
	return self.To
}

func (self Ptr2UsizeCovert) GetMut() bool {
	return false
}

func (self Ptr2UsizeCovert) IsTemporary() bool {
	return true
}

func (self Ptr2UsizeCovert) covert() {}

// Usize2PtrCovert usize转换成指针
type Usize2PtrCovert struct {
	From Expr
	To   Type
}

func (self Usize2PtrCovert) stmt() {}

func (self Usize2PtrCovert) GetType() Type {
	return self.To
}

func (self Usize2PtrCovert) GetMut() bool {
	return false
}

func (self Usize2PtrCovert) IsTemporary() bool {
	return true
}

func (self Usize2PtrCovert) covert() {}

// PtrCovert 指针转换
type PtrCovert struct {
	From Expr
	To   Type
}

func (self PtrCovert) stmt() {}

func (self PtrCovert) GetType() Type {
	return self.To
}

func (self PtrCovert) GetMut() bool {
	return false
}

func (self PtrCovert) IsTemporary() bool {
	return true
}

func (self PtrCovert) covert() {}

// UpCovert 向上转型
type UpCovert struct {
	From Expr
	To   Type
}

func (self UpCovert) stmt() {}

func (self UpCovert) GetType() Type {
	return self.To
}

func (self UpCovert) GetMut() bool {
	return false
}

func (self UpCovert) IsTemporary() bool {
	return true
}

func (self UpCovert) covert() {}

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
