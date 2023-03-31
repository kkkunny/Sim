package hir

// Expr 表达式
type Expr interface {
	Stmt
	Type() Type      // 获取类型
	Mutable() bool   // 是否可变
	Immediate() bool // 是否是立即数
}

// Ident 标识符
type Ident interface {
	Expr
	ident()
}

// Param 参数
type Param struct {
	Mut bool // 是否可变
	Typ Type // 类型
}

func NewParam(mut bool, t Type) *Param {
	return &Param{
		Mut: mut,
		Typ: t,
	}
}

func (self Param) stmt() {}

func (self Param) Type() Type {
	return self.Typ
}

func (self Param) Mutable() bool {
	return self.Mut
}

func (self Param) Immediate() bool {
	return false
}

func (self Param) ident() {}

// Boolean 布尔数
type Boolean struct {
	Typ   Type
	Value bool
}

func NewBoolean(t Type, v bool) *Boolean {
	return &Boolean{
		Typ:   t,
		Value: v,
	}
}

func (self Boolean) stmt() {}

func (self Boolean) Type() Type {
	return self.Typ
}

func (self Boolean) Mutable() bool {
	return false
}

func (self Boolean) Immediate() bool {
	return true
}

// Integer 整数
type Integer struct {
	Typ   Type
	Value int64
}

func NewInteger(t Type, v int64) *Integer {
	return &Integer{
		Typ:   t,
		Value: v,
	}
}

func (self Integer) stmt() {}

func (self Integer) Type() Type {
	return self.Typ
}

func (self Integer) Mutable() bool {
	return false
}

func (self Integer) Immediate() bool {
	return true
}

// Float 浮点数
type Float struct {
	Typ   Type
	Value float64
}

func NewFloat(t Type, v float64) *Float {
	return &Float{
		Typ:   t,
		Value: v,
	}
}

func (self Float) stmt() {}

func (self Float) Type() Type {
	return self.Typ
}

func (self Float) Mutable() bool {
	return false
}

func (self Float) Immediate() bool {
	return true
}

// String 字符串
type String struct {
	Typ   Type
	Value string
}

func NewString(t Type, v string) *String {
	return &String{
		Typ:   t,
		Value: v,
	}
}

func (self String) stmt() {}

func (self String) Type() Type {
	return self.Typ
}

func (self String) Mutable() bool {
	return false
}

func (self String) Immediate() bool {
	return true
}

// EmptyPtr 空指针
type EmptyPtr struct {
	Typ Type
}

func NewEmptyPtr(t Type) *EmptyPtr {
	return &EmptyPtr{
		Typ: t,
	}
}

func (self EmptyPtr) stmt() {}

func (self EmptyPtr) Type() Type {
	return self.Typ
}

func (self EmptyPtr) Mutable() bool {
	return false
}

func (self EmptyPtr) Immediate() bool {
	return true
}

// EmptyFunc 空函数
type EmptyFunc struct {
	Typ Type
}

func NewEmptyFunc(t Type) *EmptyFunc {
	return &EmptyFunc{
		Typ: t,
	}
}

func (self EmptyFunc) stmt() {}

func (self EmptyFunc) Type() Type {
	return self.Typ
}

func (self EmptyFunc) Mutable() bool {
	return false
}

func (self EmptyFunc) Immediate() bool {
	return true
}

// EmptyArray 空数组
type EmptyArray struct {
	Typ Type
}

func NewEmptyArray(t Type) *EmptyArray {
	return &EmptyArray{
		Typ: t,
	}
}

func (self EmptyArray) stmt() {}

func (self EmptyArray) Type() Type {
	return self.Typ
}

func (self EmptyArray) Mutable() bool {
	return false
}

func (self EmptyArray) Immediate() bool {
	return true
}

// EmptyTuple 空元组
type EmptyTuple struct {
	Typ Type
}

func NewEmptyTuple(t Type) *EmptyTuple {
	return &EmptyTuple{
		Typ: t,
	}
}

func (self EmptyTuple) stmt() {}

func (self EmptyTuple) Type() Type {
	return self.Typ
}

func (self EmptyTuple) Mutable() bool {
	return false
}

func (self EmptyTuple) Immediate() bool {
	return true
}

// EmptyStruct 空结构体
type EmptyStruct struct {
	Typ Type
}

func NewEmptyStruct(t Type) *EmptyStruct {
	return &EmptyStruct{
		Typ: t,
	}
}

func (self EmptyStruct) stmt() {}

func (self EmptyStruct) Type() Type {
	return self.Typ
}

func (self EmptyStruct) Mutable() bool {
	return false
}

func (self EmptyStruct) Immediate() bool {
	return true
}

// EmptyUnion 空联合
type EmptyUnion struct {
	Typ Type
}

func NewEmptyUnion(t Type) *EmptyUnion {
	return &EmptyUnion{
		Typ: t,
	}
}

func (self EmptyUnion) stmt() {}

func (self EmptyUnion) Type() Type {
	return self.Typ
}

func (self EmptyUnion) Mutable() bool {
	return false
}

func (self EmptyUnion) Immediate() bool {
	return true
}

// Array 数组
type Array struct {
	Typ   Type
	Elems []Expr
}

func NewArray(t Type, elem ...Expr) *Array {
	return &Array{
		Typ:   t,
		Elems: elem,
	}
}

func (self Array) stmt() {}

func (self Array) Type() Type {
	return self.Typ
}

func (self Array) Mutable() bool {
	return false
}

func (self Array) Immediate() bool {
	return true
}

// Tuple 元组
type Tuple struct {
	Typ   Type
	Elems []Expr
}

func NewTuple(t Type, elem ...Expr) *Tuple {
	return &Tuple{
		Typ:   t,
		Elems: elem,
	}
}

func (self Tuple) stmt() {}

func (self Tuple) Type() Type {
	return self.Typ
}

func (self Tuple) Mutable() bool {
	return false
}

func (self Tuple) Immediate() bool {
	return true
}

// Struct 结构体
type Struct struct {
	Typ    Type
	Fields []Expr
}

func NewStruct(t Type, field ...Expr) *Struct {
	return &Struct{
		Typ:    t,
		Fields: field,
	}
}

func (self Struct) stmt() {}

func (self Struct) Type() Type {
	return self.Typ
}

func (self Struct) Mutable() bool {
	return false
}

func (self Struct) Immediate() bool {
	return true
}

// Unary 一元表达式
type Unary interface {
	Expr
	unary()
}

// Not 布尔取反
type Not struct {
	Value Expr
}

func NewNot(v Expr) *Not {
	return &Not{
		Value: v,
	}
}

func (self Not) stmt() {}

func (self Not) Type() Type {
	return self.Value.Type()
}

func (self Not) Mutable() bool {
	return false
}

func (self Not) Immediate() bool {
	return true
}

func (self Not) unary() {}

// GetPointer 获取指针
type GetPointer struct {
	Value Expr
}

func NewGetPointer(v Expr) *GetPointer {
	return &GetPointer{
		Value: v,
	}
}

func (self GetPointer) stmt() {}

func (self GetPointer) Type() Type {
	return NewTypePtr(self.Value.Type())
}

func (self GetPointer) Mutable() bool {
	return false
}

func (self GetPointer) Immediate() bool {
	return true
}

func (self GetPointer) unary() {}

// GetValue 获取值
type GetValue struct {
	Value Expr
}

func NewGetValue(v Expr) *GetValue {
	return &GetValue{
		Value: v,
	}
}

func (self GetValue) stmt() {}

func (self GetValue) Type() Type {
	return self.Value.Type().GetPtr()
}

func (self GetValue) Mutable() bool {
	return self.Value.Mutable()
}

func (self GetValue) Immediate() bool {
	return self.Value.Immediate()
}

func (self GetValue) unary() {}

// Binary 二元表达式
type Binary interface {
	Expr
	binary()
}

// Add 加
type Add struct {
	Left, Right Expr
}

func NewAdd(l, r Expr) *Add {
	return &Add{
		Left:  l,
		Right: r,
	}
}

func (self Add) stmt() {}

func (self Add) Type() Type {
	return self.Left.Type()
}

func (self Add) Mutable() bool {
	return false
}

func (self Add) Immediate() bool {
	return true
}

func (self Add) binary() {}

// Sub 减
type Sub struct {
	Left, Right Expr
}

func NewSub(l, r Expr) *Sub {
	return &Sub{
		Left:  l,
		Right: r,
	}
}

func (self Sub) stmt() {}

func (self Sub) Type() Type {
	return self.Left.Type()
}

func (self Sub) Mutable() bool {
	return false
}

func (self Sub) Immediate() bool {
	return true
}

func (self Sub) binary() {}

// Mul 乘
type Mul struct {
	Left, Right Expr
}

func NewMul(l, r Expr) *Mul {
	return &Mul{
		Left:  l,
		Right: r,
	}
}

func (self Mul) stmt() {}

func (self Mul) Type() Type {
	return self.Left.Type()
}

func (self Mul) Mutable() bool {
	return false
}

func (self Mul) Immediate() bool {
	return true
}

func (self Mul) binary() {}

// Div 除
type Div struct {
	Left, Right Expr
}

func NewDiv(l, r Expr) *Div {
	return &Div{
		Left:  l,
		Right: r,
	}
}

func (self Div) stmt() {}

func (self Div) Type() Type {
	return self.Left.Type()
}

func (self Div) Mutable() bool {
	return false
}

func (self Div) Immediate() bool {
	return true
}

func (self Div) binary() {}

// Mod 取余
type Mod struct {
	Left, Right Expr
}

func NewMod(l, r Expr) *Mod {
	return &Mod{
		Left:  l,
		Right: r,
	}
}

func (self Mod) stmt() {}

func (self Mod) Type() Type {
	return self.Left.Type()
}

func (self Mod) Mutable() bool {
	return false
}

func (self Mod) Immediate() bool {
	return true
}

func (self Mod) binary() {}

// And 与
type And struct {
	Left, Right Expr
}

func NewAnd(l, r Expr) *And {
	return &And{
		Left:  l,
		Right: r,
	}
}

func (self And) stmt() {}

func (self And) Type() Type {
	return self.Left.Type()
}

func (self And) Mutable() bool {
	return false
}

func (self And) Immediate() bool {
	return true
}

func (self And) binary() {}

// Or 或
type Or struct {
	Left, Right Expr
}

func NewOr(l, r Expr) *Or {
	return &Or{
		Left:  l,
		Right: r,
	}
}

func (self Or) stmt() {}

func (self Or) Type() Type {
	return self.Left.Type()
}

func (self Or) Mutable() bool {
	return false
}

func (self Or) Immediate() bool {
	return true
}

func (self Or) binary() {}

// Xor 异或
type Xor struct {
	Left, Right Expr
}

func NewXor(l, r Expr) *Xor {
	return &Xor{
		Left:  l,
		Right: r,
	}
}

func (self Xor) stmt() {}

func (self Xor) Type() Type {
	return self.Left.Type()
}

func (self Xor) Mutable() bool {
	return false
}

func (self Xor) Immediate() bool {
	return true
}

func (self Xor) binary() {}

// Shl 左移
type Shl struct {
	Left, Right Expr
}

func NewShl(l, r Expr) *Shl {
	return &Shl{
		Left:  l,
		Right: r,
	}
}

func (self Shl) stmt() {}

func (self Shl) Type() Type {
	return self.Left.Type()
}

func (self Shl) Mutable() bool {
	return false
}

func (self Shl) Immediate() bool {
	return true
}

func (self Shl) binary() {}

// Shr 右移
type Shr struct {
	Left, Right Expr
}

func NewShr(l, r Expr) *Shr {
	return &Shr{
		Left:  l,
		Right: r,
	}
}

func (self Shr) stmt() {}

func (self Shr) Type() Type {
	return self.Left.Type()
}

func (self Shr) Mutable() bool {
	return false
}

func (self Shr) Immediate() bool {
	return true
}

func (self Shr) binary() {}

// LogicAnd 逻辑与
type LogicAnd struct {
	Left, Right Expr
}

func NewLogicAnd(l, r Expr) *LogicAnd {
	return &LogicAnd{
		Left:  l,
		Right: r,
	}
}

func (self LogicAnd) stmt() {}

func (self LogicAnd) Type() Type {
	return self.Left.Type()
}

func (self LogicAnd) Mutable() bool {
	return false
}

func (self LogicAnd) Immediate() bool {
	return true
}

func (self LogicAnd) binary() {}

// LogicOr 逻辑或
type LogicOr struct {
	Left, Right Expr
}

func NewLogicOr(l, r Expr) *LogicOr {
	return &LogicOr{
		Left:  l,
		Right: r,
	}
}

func (self LogicOr) stmt() {}

func (self LogicOr) Type() Type {
	return self.Left.Type()
}

func (self LogicOr) Mutable() bool {
	return false
}

func (self LogicOr) Immediate() bool {
	return true
}

func (self LogicOr) binary() {}

// Equal 相等
type Equal struct {
	Typ         Type
	Left, Right Expr
}

func NewEqual(t Type, l, r Expr) *Equal {
	return &Equal{
		Typ:   t,
		Left:  l,
		Right: r,
	}
}

func (self Equal) stmt() {}

func (self Equal) Type() Type {
	return self.Typ
}

func (self Equal) Mutable() bool {
	return false
}

func (self Equal) Immediate() bool {
	return true
}

func (self Equal) binary() {}

// NotEqual 不相等
type NotEqual struct {
	Typ         Type
	Left, Right Expr
}

func NewNotEqual(t Type, l, r Expr) *NotEqual {
	return &NotEqual{
		Typ:   t,
		Left:  l,
		Right: r,
	}
}

func (self NotEqual) stmt() {}

func (self NotEqual) Type() Type {
	return self.Typ
}

func (self NotEqual) Mutable() bool {
	return false
}

func (self NotEqual) Immediate() bool {
	return true
}

func (self NotEqual) binary() {}

// Lt 小于
type Lt struct {
	Typ         Type
	Left, Right Expr
}

func NewLt(t Type, l, r Expr) *Lt {
	return &Lt{
		Typ:   t,
		Left:  l,
		Right: r,
	}
}

func (self Lt) stmt() {}

func (self Lt) Type() Type {
	return self.Typ
}

func (self Lt) Mutable() bool {
	return false
}

func (self Lt) Immediate() bool {
	return true
}

func (self Lt) binary() {}

// Le 小于等于
type Le struct {
	Typ         Type
	Left, Right Expr
}

func NewLe(t Type, l, r Expr) *Le {
	return &Le{
		Typ:   t,
		Left:  l,
		Right: r,
	}
}

func (self Le) stmt() {}

func (self Le) Type() Type {
	return self.Typ
}

func (self Le) Mutable() bool {
	return false
}

func (self Le) Immediate() bool {
	return true
}

func (self Le) binary() {}

// Assign 赋值
type Assign struct {
	Left, Right Expr
}

func NewAssign(l, r Expr) *Assign {
	return &Assign{
		Left:  l,
		Right: r,
	}
}

func (self Assign) stmt() {}

func (self Assign) Type() Type {
	return NewTypeNone()
}

func (self Assign) Mutable() bool {
	return false
}

func (self Assign) Immediate() bool {
	return true
}

func (self Assign) binary() {}

// Ternary 三元表达式
type Ternary struct {
	Cond, True, False Expr
}

func NewTernary(c, t, f Expr) *Ternary {
	return &Ternary{
		Cond:  c,
		True:  t,
		False: f,
	}
}

func (self Ternary) stmt() {}

func (self Ternary) Type() Type {
	return self.True.Type()
}

func (self Ternary) Mutable() bool {
	return false
}

func (self Ternary) Immediate() bool {
	return true
}

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

func NewFuncCall(f Expr, arg ...Expr) *FuncCall {
	return &FuncCall{
		Func: f,
		Args: arg,
	}
}

func (self FuncCall) stmt() {}

func (self FuncCall) Type() Type {
	return self.Func.Type().GetFuncRet()
}

func (self FuncCall) Mutable() bool {
	return false
}

func (self FuncCall) Immediate() bool {
	return true
}

func (self FuncCall) call() {}

// MethodCall 方法调用
type MethodCall struct {
	Method *Method
	Self   Expr
	Args   []Expr
}

func NewMethodCall(m *Method, s Expr, arg ...Expr) *MethodCall {
	return &MethodCall{
		Method: m,
		Self:   s,
		Args:   arg,
	}
}

func (self MethodCall) stmt() {}

func (self MethodCall) Type() Type {
	return self.Method.Type().GetFuncRet()
}

func (self MethodCall) Mutable() bool {
	return false
}

func (self MethodCall) Immediate() bool {
	return true
}

func (self MethodCall) call() {}

// Index 索引
type Index interface {
	Expr
	index()
}

// ArrayIndex 数组索引
type ArrayIndex struct {
	From, Index Expr
}

func NewArrayIndex(f, i Expr) *ArrayIndex {
	return &ArrayIndex{
		From:  f,
		Index: i,
	}
}

func (self ArrayIndex) stmt() {}

func (self ArrayIndex) Type() Type {
	return self.From.Type().GetArrayElem()
}

func (self ArrayIndex) Mutable() bool {
	return self.From.Mutable()
}

func (self ArrayIndex) Immediate() bool {
	return self.From.Immediate()
}

func (self ArrayIndex) index() {}

// TupleIndex 元组索引
type TupleIndex struct {
	From  Expr
	Index uint
}

func NewTupleIndex(f Expr, i uint) *TupleIndex {
	return &TupleIndex{
		From:  f,
		Index: i,
	}
}

func (self TupleIndex) stmt() {}

func (self TupleIndex) Type() Type {
	return self.From.Type().GetTupleElems()[self.Index]
}

func (self TupleIndex) Mutable() bool {
	return self.From.Mutable()
}

func (self TupleIndex) Immediate() bool {
	return self.From.Immediate()
}

func (self TupleIndex) index() {}

// PointerIndex 指针索引
type PointerIndex struct {
	From, Index Expr
}

func NewPointerIndex(f, i Expr) *PointerIndex {
	return &PointerIndex{
		From:  f,
		Index: i,
	}
}

func (self PointerIndex) stmt() {}

func (self PointerIndex) Type() Type {
	return self.From.Type().GetPtr()
}

func (self PointerIndex) Mutable() bool {
	return self.From.Mutable()
}

func (self PointerIndex) Immediate() bool {
	return self.From.Immediate()
}

func (self PointerIndex) index() {}

// GetStructField 获取结构体字段
type GetStructField struct {
	From Expr
	Attr string
}

func NewGetStructField(f Expr, a string) *GetStructField {
	return &GetStructField{
		From: f,
		Attr: a,
	}
}

func (self GetStructField) stmt() {}

func (self GetStructField) Type() Type {
	for _, f := range self.From.Type().GetStructFields() {
		if f.Second == self.Attr {
			return f.Third
		}
	}
	panic("unreachable")
}

func (self GetStructField) Mutable() bool {
	return self.From.Mutable()
}

func (self GetStructField) Immediate() bool {
	return self.From.Immediate()
}

func (self GetStructField) GetFieldIndex() uint {
	for i, f := range self.From.Type().GetStructFields() {
		if f.Second == self.Attr {
			return uint(i)
		}
	}
	panic("unreachable")
}

// Covert 类型转换
type Covert interface {
	Expr
	GetFrom() Expr
	covert()
}

// Int2Int 整型转整型
type Int2Int struct {
	From Expr
	To   Type
}

func NewInt2Int(f Expr, t Type) *Int2Int {
	return &Int2Int{
		From: f,
		To:   t,
	}
}

func (self Int2Int) stmt() {}

func (self Int2Int) Type() Type {
	return self.To
}

func (self Int2Int) Mutable() bool {
	return false
}

func (self Int2Int) Immediate() bool {
	return true
}

func (self Int2Int) covert() {}

func (self Int2Int) GetFrom() Expr {
	return self.From
}

// Float2Float 浮点型转浮点型
type Float2Float struct {
	From Expr
	To   Type
}

func NewFloat2Float(f Expr, t Type) *Float2Float {
	return &Float2Float{
		From: f,
		To:   t,
	}
}

func (self Float2Float) stmt() {}

func (self Float2Float) Type() Type {
	return self.To
}

func (self Float2Float) Mutable() bool {
	return false
}

func (self Float2Float) Immediate() bool {
	return true
}

func (self Float2Float) covert() {}

func (self Float2Float) GetFrom() Expr {
	return self.From
}

// Int2Float 整型转浮点型
type Int2Float struct {
	From Expr
	To   Type
}

func NewInt2Float(f Expr, t Type) *Int2Float {
	return &Int2Float{
		From: f,
		To:   t,
	}
}

func (self Int2Float) stmt() {}

func (self Int2Float) Type() Type {
	return self.To
}

func (self Int2Float) Mutable() bool {
	return false
}

func (self Int2Float) Immediate() bool {
	return true
}

func (self Int2Float) covert() {}

func (self Int2Float) GetFrom() Expr {
	return self.From
}

// Float2Int 浮点型转整型
type Float2Int struct {
	From Expr
	To   Type
}

func NewFloat2Int(f Expr, t Type) *Float2Int {
	return &Float2Int{
		From: f,
		To:   t,
	}
}

func (self Float2Int) stmt() {}

func (self Float2Int) Type() Type {
	return self.To
}

func (self Float2Int) Mutable() bool {
	return false
}

func (self Float2Int) Immediate() bool {
	return true
}

func (self Float2Int) covert() {}

func (self Float2Int) GetFrom() Expr {
	return self.From
}

// Ptr2Ptr 指针转指针
type Ptr2Ptr struct {
	From Expr
	To   Type
}

func NewPtr2Ptr(f Expr, t Type) *Ptr2Ptr {
	return &Ptr2Ptr{
		From: f,
		To:   t,
	}
}

func (self Ptr2Ptr) stmt() {}

func (self Ptr2Ptr) Type() Type {
	return self.To
}

func (self Ptr2Ptr) Mutable() bool {
	return false
}

func (self Ptr2Ptr) Immediate() bool {
	return true
}

func (self Ptr2Ptr) covert() {}

func (self Ptr2Ptr) GetFrom() Expr {
	return self.From
}

// Usize2Ptr usize转指针
type Usize2Ptr struct {
	From Expr
	To   Type
}

func NewUsize2Ptr(f Expr, t Type) *Usize2Ptr {
	return &Usize2Ptr{
		From: f,
		To:   t,
	}
}

func (self Usize2Ptr) stmt() {}

func (self Usize2Ptr) Type() Type {
	return self.To
}

func (self Usize2Ptr) Mutable() bool {
	return false
}

func (self Usize2Ptr) Immediate() bool {
	return true
}

func (self Usize2Ptr) covert() {}

func (self Usize2Ptr) GetFrom() Expr {
	return self.From
}

// Ptr2Usize 指针转usize
type Ptr2Usize struct {
	From Expr
	To   Type
}

func NewPtr2Usize(f Expr, t Type) *Ptr2Usize {
	return &Ptr2Usize{
		From: f,
		To:   t,
	}
}

func (self Ptr2Usize) stmt() {}

func (self Ptr2Usize) Type() Type {
	return self.To
}

func (self Ptr2Usize) Mutable() bool {
	return false
}

func (self Ptr2Usize) Immediate() bool {
	return true
}

func (self Ptr2Usize) covert() {}

func (self Ptr2Usize) GetFrom() Expr {
	return self.From
}

// WrapCovert 包装转换
type WrapCovert struct {
	From Expr
	To   Type
}

func NewWrapCovert(f Expr, t Type) *WrapCovert {
	return &WrapCovert{
		From: f,
		To:   t,
	}
}

func (self WrapCovert) stmt() {}

func (self WrapCovert) Type() Type {
	return self.To
}

func (self WrapCovert) Mutable() bool {
	return false
}

func (self WrapCovert) Immediate() bool {
	return true
}

func (self WrapCovert) covert() {}

func (self WrapCovert) GetFrom() Expr {
	return self.From
}

// WrapUnion 包装联合
type WrapUnion struct {
	From Expr
	To   Type
}

func NewWrapUnion(f Expr, t Type) *WrapUnion {
	return &WrapUnion{
		From: f,
		To:   t,
	}
}

func (self WrapUnion) stmt() {}

func (self WrapUnion) Type() Type {
	return self.To
}

func (self WrapUnion) Mutable() bool {
	return false
}

func (self WrapUnion) Immediate() bool {
	return true
}

func (self WrapUnion) covert() {}

func (self WrapUnion) GetFrom() Expr {
	return self.From
}

// UnwrapUnion 解包装联合
type UnwrapUnion struct {
	From Expr
	To   Type
}

func NewUnwrapUnion(f Expr, t Type) *UnwrapUnion {
	return &UnwrapUnion{
		From: f,
		To:   t,
	}
}

func (self UnwrapUnion) stmt() {}

func (self UnwrapUnion) Type() Type {
	return self.To
}

func (self UnwrapUnion) Mutable() bool {
	return false
}

func (self UnwrapUnion) Immediate() bool {
	return true
}

func (self UnwrapUnion) covert() {}

func (self UnwrapUnion) GetFrom() Expr {
	return self.From
}

// Size 获取类型大小
type Size struct {
	Typ Type
}

func NewSize(t Type) *Size {
	return &Size{
		Typ: t,
	}
}

func (self Size) stmt() {}

func (self Size) Type() Type {
	return NewTypeUsize()
}

func (self Size) Mutable() bool {
	return false
}

func (self Size) Immediate() bool {
	return true
}

// Align 获取类型对齐大小
type Align struct {
	Typ Type
}

func NewAlign(t Type) *Align {
	return &Align{
		Typ: t,
	}
}

func (self Align) stmt() {}

func (self Align) Type() Type {
	return NewTypeUsize()
}

func (self Align) Mutable() bool {
	return false
}

func (self Align) Immediate() bool {
	return true
}
