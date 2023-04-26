package mir

// Inst 语句
type Inst interface {
	inst()
}

// Alloc 栈空间分配
type Alloc struct {
	Type Type
}

func (self *Block) NewAlloc(t Type) *Alloc {
	inst := &Alloc{
		Type: t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Alloc) inst() {}
func (self Alloc) GetType() Type {
	return NewTypePtr(self.Type)
}

// Load 获取值
type Load struct {
	Value Value
}

func (self *Block) NewLoad(v Value) *Load {
	if !v.GetType().IsPtr() {
		panic("unreachable")
	}
	inst := &Load{
		Value: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Load) inst() {}
func (self Load) GetType() Type {
	return self.Value.GetType().GetPtr()
}

// Store 赋值
type Store struct {
	From, To Value
}

func (self *Block) NewStore(f, t Value) *Store {
	if !t.GetType().IsPtr() || !f.GetType().Equal(t.GetType().GetPtr()) {
		panic("unreachable")
	}
	inst := &Store{From: f, To: t}
	self.Insts.PushBack(inst)
	return inst
}
func (self Store) inst() {}

// Unreachable 不可到达
type Unreachable struct{}

func (self *Block) NewUnreachable() *Unreachable {
	inst := &Unreachable{}
	self.Insts.PushBack(inst)
	return inst
}
func (self Unreachable) inst() {}

// Return 函数返回
type Return struct {
	Value Value // 可能为空
}

func (self *Block) NewReturn(v Value) *Return {
	retType := self.Belong.Type.GetFuncRet()
	if (v == nil && !retType.IsVoid()) ||
		(v != nil && !v.GetType().Equal(self.Belong.Type.GetFuncRet())) {
		panic("unreachable")
	}
	inst := &Return{v}
	self.Insts.PushBack(inst)
	return inst
}
func (self Return) inst() {}

// Jmp 跳转
type Jmp struct {
	Dst *Block
}

func (self *Block) NewJmp(dst *Block) *Jmp {
	inst := &Jmp{Dst: dst}
	self.Insts.PushBack(inst)
	return inst
}
func (self Jmp) inst() {}

// CondJmp 条件跳转
type CondJmp struct {
	Cond        Value
	True, False *Block
}

func (self *Block) NewCondJmp(c Value, t, f *Block) *CondJmp {
	inst := &CondJmp{
		Cond:  c,
		True:  t,
		False: f,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self CondJmp) inst() {}

// Eq 相等比较
type Eq struct {
	Left, Right Value
}

func (self *Block) NewEq(l, r Value) *Eq {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || (!t.IsNumber() && !t.IsPtr() && !t.IsFunc() && !t.IsBool()) {
		panic("unreachable")
	}
	inst := &Eq{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Eq) inst() {}
func (self Eq) GetType() Type {
	return NewTypeBool()
}

// Ne 不等比较
type Ne struct {
	Left, Right Value
}

func (self *Block) NewNe(l, r Value) *Ne {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || (!t.IsNumber() && !t.IsPtr() && !t.IsFunc() && !t.IsBool()) {
		panic("unreachable")
	}
	inst := &Ne{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Ne) inst() {}
func (self Ne) GetType() Type {
	return NewTypeBool()
}

// Lt 小于
type Lt struct {
	Left, Right Value
}

func (self *Block) NewLt(l, r Value) *Lt {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || !t.IsNumber() {
		panic("unreachable")
	}
	inst := &Lt{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Lt) inst() {}
func (self Lt) GetType() Type {
	return NewTypeBool()
}

// Le 小于等于
type Le struct {
	Left, Right Value
}

func (self *Block) NewLe(l, r Value) *Le {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || !t.IsNumber() {
		panic("unreachable")
	}
	inst := &Le{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Le) inst() {}
func (self Le) GetType() Type {
	return NewTypeBool()
}

// Gt 大于
type Gt struct {
	Left, Right Value
}

func (self *Block) NewGt(l, r Value) *Gt {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || !t.IsNumber() {
		panic("unreachable")
	}
	inst := &Gt{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Gt) inst() {}
func (self Gt) GetType() Type {
	return NewTypeBool()
}

// Ge 大于等于
type Ge struct {
	Left, Right Value
}

func (self *Block) NewGe(l, r Value) *Ge {
	t := l.GetType()
	if !l.GetType().Equal(r.GetType()) || !t.IsNumber() {
		panic("unreachable")
	}
	inst := &Ge{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Ge) inst() {}
func (self Ge) GetType() Type {
	return NewTypeBool()
}

// ArrayIndex 数组索引
type ArrayIndex struct {
	From, Index Value
}

func (self *Block) NewArrayIndex(f, i Value) *ArrayIndex {
	if !i.GetType().IsUint() || i.GetType().GetWidth() != 0 {
		panic("unreachable")
	} else if !f.GetType().IsArray() && !(f.GetType().IsPtr() && f.GetType().GetPtr().IsArray()) {
		panic("unreachable")
	}
	inst := &ArrayIndex{
		From:  f,
		Index: i,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self ArrayIndex) inst() {}
func (self ArrayIndex) GetType() Type {
	ft := self.From.GetType()
	if ft.IsPtr() {
		return NewTypePtr(ft.GetPtr().GetArrayElem())
	}
	return ft.GetArrayElem()
}
func (self ArrayIndex) IsPtr() bool {
	if self.From.GetType().IsPtr() {
		return true
	}
	return false
}

// StructIndex 结构体索引
type StructIndex struct {
	From  Value
	Index uint
}

func (self *Block) NewStructIndex(f Value, i uint) *StructIndex {
	ft := f.GetType()
	if !ft.IsStruct() && !(ft.IsPtr() && ft.GetPtr().IsStruct()) {
		panic("unreachable")
	}
	var elemCount int
	if f.GetType().IsPtr() {
		elemCount = len(f.GetType().GetPtr().GetStructElems())
	} else {
		elemCount = len(f.GetType().GetStructElems())
	}
	if i >= uint(elemCount) {
		panic("unreachable")
	}
	inst := &StructIndex{
		From:  f,
		Index: i,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self StructIndex) inst() {}
func (self StructIndex) GetType() Type {
	ft := self.From.GetType()
	if ft.IsPtr() {
		return NewTypePtr(ft.GetPtr().GetStructElems()[self.Index])
	}
	return ft.GetStructElems()[self.Index]
}
func (self StructIndex) IsPtr() bool {
	if self.From.GetType().IsPtr() {
		return true
	}
	return false
}

// PointerIndex 指针索引
type PointerIndex struct {
	From, Index Value
}

func (self *Block) NewPointerIndex(f, i Value) *PointerIndex {
	if !i.GetType().IsUint() || i.GetType().GetWidth() != 0 || !f.GetType().IsPtr() {
		panic("unreachable")
	}
	inst := &PointerIndex{
		From:  f,
		Index: i,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self PointerIndex) inst() {}
func (self PointerIndex) GetType() Type {
	return self.From.GetType()
}

// UnwrapUnion 取出联合
type UnwrapUnion struct {
	To    Type
	Value Value
}

func (self *Block) NewUnwrapUnion(t Type, v Value) *UnwrapUnion {
	vt := v.GetType()
	if !vt.IsUnion() && vt.IsPtr() && !vt.GetPtr().IsUnion() {
		panic("unreachable")
	}
	var elems []Type
	if vt.IsPtr() {
		elems = vt.GetPtr().GetUnionElems()
	} else {
		elems = vt.GetUnionElems()
	}
	var exist bool
	for _, e := range elems {
		if t.Equal(e) {
			exist = true
			break
		}
	}
	if !exist {
		panic("unreachable")
	}

	inst := &UnwrapUnion{
		To:    t,
		Value: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self UnwrapUnion) inst() {}
func (self UnwrapUnion) GetType() Type {
	if self.IsPtr() {
		return NewTypePtr(self.To)
	}
	return self.To
}
func (self UnwrapUnion) IsPtr() bool {
	return self.Value.GetType().IsPtr()
}

// Add 加
type Add struct {
	Left, Right Value
}

func (self *Block) NewAdd(l, r Value) *Add {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsNumber() {
		panic("unreachable")
	}
	inst := &Add{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Add) inst() {}
func (self Add) GetType() Type {
	return self.Left.GetType()
}

// Sub 减
type Sub struct {
	Left, Right Value
}

func (self *Block) NewSub(l, r Value) *Sub {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsNumber() {
		panic("unreachable")
	}
	inst := &Sub{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Sub) inst() {}
func (self Sub) GetType() Type {
	return self.Left.GetType()
}

// Mul 乘
type Mul struct {
	Left, Right Value
}

func (self *Block) NewMul(l, r Value) *Mul {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsNumber() {
		panic("unreachable")
	}
	inst := &Mul{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Mul) inst() {}
func (self Mul) GetType() Type {
	return self.Left.GetType()
}

// Div 除
type Div struct {
	Left, Right Value
}

func (self *Block) NewDiv(l, r Value) *Div {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsNumber() {
		panic("unreachable")
	}
	inst := &Div{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Div) inst() {}
func (self Div) GetType() Type {
	return self.Left.GetType()
}

// Mod 取于
type Mod struct {
	Left, Right Value
}

func (self *Block) NewMod(l, r Value) *Mod {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsNumber() {
		panic("unreachable")
	}
	inst := &Mod{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Mod) inst() {}
func (self Mod) GetType() Type {
	return self.Left.GetType()
}

// And 与
type And struct {
	Left, Right Value
}

func (self *Block) NewAnd(l, r Value) *And {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &And{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self And) inst() {}
func (self And) GetType() Type {
	return self.Left.GetType()
}

// Or 或
type Or struct {
	Left, Right Value
}

func (self *Block) NewOr(l, r Value) *Or {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &Or{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Or) inst() {}
func (self Or) GetType() Type {
	return self.Left.GetType()
}

// Xor 异或
type Xor struct {
	Left, Right Value
}

func (self *Block) NewXor(l, r Value) *Xor {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &Xor{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Xor) inst() {}
func (self Xor) GetType() Type {
	return self.Left.GetType()
}

// Shl 左移
type Shl struct {
	Left, Right Value
}

func (self *Block) NewShl(l, r Value) *Shl {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &Shl{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Shl) inst() {}
func (self Shl) GetType() Type {
	return self.Left.GetType()
}

// Shr 右移
type Shr struct {
	Left, Right Value
}

func (self *Block) NewShr(l, r Value) *Shr {
	if !l.GetType().Equal(r.GetType()) || !l.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &Shr{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Shr) inst() {}
func (self Shr) GetType() Type {
	return self.Left.GetType()
}

// Phi phi
type Phi struct {
	Froms  []*Block
	Values []Value
}

func (self *Block) NewPhi(f []*Block, v []Value) *Phi {
	if len(f) != len(v) || len(f) == 0 {
		panic("unreachable")
	}
	vt := v[0].GetType()
	for _, vv := range v[1:] {
		if !vv.GetType().Equal(vt) {
			panic("unreachable")
		}
	}
	inst := &Phi{
		Froms:  f,
		Values: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Phi) inst() {}
func (self Phi) GetType() Type {
	return self.Values[0].GetType()
}

// Sint2Sint sint -> sint
type Sint2Sint struct {
	From Value
	To   Type
}

func (self *Block) NewSint2Sint(f Value, t Type) *Sint2Sint {
	if !f.GetType().IsSint() || !t.IsSint() {
		panic("unreachable")
	}
	inst := &Sint2Sint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Sint2Sint) inst() {}
func (self Sint2Sint) GetType() Type {
	return self.To
}

// Sint2Uint sint -> uint
type Sint2Uint struct {
	From Value
	To   Type
}

func (self *Block) NewSint2Uint(f Value, t Type) *Sint2Uint {
	if !f.GetType().IsSint() || !t.IsUint() {
		panic("unreachable")
	}
	inst := &Sint2Uint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Sint2Uint) inst() {}
func (self Sint2Uint) GetType() Type {
	return self.To
}

// Sint2Float sint -> flaot
type Sint2Float struct {
	From Value
	To   Type
}

func (self *Block) NewSint2Float(f Value, t Type) *Sint2Float {
	if !f.GetType().IsSint() || !t.IsFloat() {
		panic("unreachable")
	}
	inst := &Sint2Float{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Sint2Float) inst() {}
func (self Sint2Float) GetType() Type {
	return self.To
}

// Sint2Ptr sint -> ptr
type Sint2Ptr struct {
	From Value
	To   Type
}

func (self *Block) NewSint2Ptr(f Value, t Type) *Sint2Ptr {
	if !f.GetType().IsSint() || !t.IsPtr() {
		panic("unreachable")
	}
	inst := &Sint2Ptr{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Sint2Ptr) inst() {}
func (self Sint2Ptr) GetType() Type {
	return self.To
}

// Uint2Uint uint -> uint
type Uint2Uint struct {
	From Value
	To   Type
}

func (self *Block) NewUint2Uint(f Value, t Type) *Uint2Uint {
	if !f.GetType().IsUint() || !t.IsUint() {
		panic("unreachable")
	}
	inst := &Uint2Uint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Uint2Uint) inst() {}
func (self Uint2Uint) GetType() Type {
	return self.To
}

// Uint2Sint uint -> sint
type Uint2Sint struct {
	From Value
	To   Type
}

func (self *Block) NewUint2Sint(f Value, t Type) *Uint2Sint {
	if !f.GetType().IsUint() || !t.IsSint() {
		panic("unreachable")
	}
	inst := &Uint2Sint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Uint2Sint) inst() {}
func (self Uint2Sint) GetType() Type {
	return self.To
}

// Uint2Float uint -> float
type Uint2Float struct {
	From Value
	To   Type
}

func (self *Block) NewUint2Float(f Value, t Type) *Uint2Float {
	if !f.GetType().IsUint() || !t.IsFloat() {
		panic("unreachable")
	}
	inst := &Uint2Float{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Uint2Float) inst() {}
func (self Uint2Float) GetType() Type {
	return self.To
}

// Uint2Ptr uint -> ptr
type Uint2Ptr struct {
	From Value
	To   Type
}

func (self *Block) NewUint2Ptr(f Value, t Type) *Uint2Ptr {
	if !f.GetType().IsUint() || !t.IsPtr() {
		panic("unreachable")
	}
	inst := &Uint2Ptr{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Uint2Ptr) inst() {}
func (self Uint2Ptr) GetType() Type {
	return self.To
}

// Float2Sint flaot -> sint
type Float2Sint struct {
	From Value
	To   Type
}

func (self *Block) NewFloat2Sint(f Value, t Type) *Float2Sint {
	if !f.GetType().IsFloat() || !t.IsSint() {
		panic("unreachable")
	}
	inst := &Float2Sint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Float2Sint) inst() {}
func (self Float2Sint) GetType() Type {
	return self.To
}

// Float2Uint flaot -> uint
type Float2Uint struct {
	From Value
	To   Type
}

func (self *Block) NewFloat2Uint(f Value, t Type) *Float2Uint {
	if !f.GetType().IsFloat() || !t.IsUint() {
		panic("unreachable")
	}
	inst := &Float2Uint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Float2Uint) inst() {}
func (self Float2Uint) GetType() Type {
	return self.To
}

// Float2Float flaot -> float
type Float2Float struct {
	From Value
	To   Type
}

func (self *Block) NewFloat2Float(f Value, t Type) *Float2Float {
	if !f.GetType().IsFloat() || !t.IsUint() {
		panic("unreachable")
	}
	inst := &Float2Float{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Float2Float) inst() {}
func (self Float2Float) GetType() Type {
	return self.To
}

// Ptr2Ptr ptr -> ptr
type Ptr2Ptr struct {
	From Value
	To   Type
}

func (self *Block) NewPtr2Ptr(f Value, t Type) *Ptr2Ptr {
	if !f.GetType().IsPtr() || !t.IsPtr() {
		panic("unreachable")
	}
	inst := &Ptr2Ptr{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Ptr2Ptr) inst() {}
func (self Ptr2Ptr) GetType() Type {
	return self.To
}

// Ptr2Sint ptr -> sint
type Ptr2Sint struct {
	From Value
	To   Type
}

func (self *Block) NewPtr2Sint(f Value, t Type) *Ptr2Sint {
	if !f.GetType().IsPtr() || !t.IsSint() {
		panic("unreachable")
	}
	inst := &Ptr2Sint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Ptr2Sint) inst() {}
func (self Ptr2Sint) GetType() Type {
	return self.To
}

// Ptr2Uint ptr -> uint
type Ptr2Uint struct {
	From Value
	To   Type
}

func (self *Block) NewPtr2Uint(f Value, t Type) *Ptr2Uint {
	if !f.GetType().IsPtr() || !t.IsUint() {
		panic("unreachable")
	}
	inst := &Ptr2Uint{
		From: f,
		To:   t,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Ptr2Uint) inst() {}
func (self Ptr2Uint) GetType() Type {
	return self.To
}

// Select 选择
type Select struct {
	Cond, True, False Value
}

func (self *Block) NewSelect(c, t, f Value) *Select {
	if !c.GetType().IsBool() || !t.GetType().Equal(f.GetType()) {
		panic("unreachable")
	}
	inst := &Select{
		Cond:  c,
		True:  t,
		False: f,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Select) inst() {}
func (self Select) GetType() Type {
	return self.True.GetType()
}

// Not 取反
type Not struct {
	Value Value
}

func (self *Block) NewNot(v Value) *Not {
	if !v.GetType().IsInteger() {
		panic("unreachable")
	}
	inst := &Not{
		Value: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Not) inst() {}
func (self Not) GetType() Type {
	return self.Value.GetType()
}

// LogicNot 逻辑取反
type LogicNot struct {
	Value Value
}

func (self *Block) NewLogicNot(v Value) *LogicNot {
	if !v.GetType().IsBool() {
		panic("unreachable")
	}
	inst := &LogicNot{
		Value: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self LogicNot) inst() {}
func (self LogicNot) GetType() Type {
	return self.Value.GetType()
}

// LogicAnd 逻辑与
type LogicAnd struct {
	Left, Right Value
}

func (self *Block) NewLogicAnd(l, r Value) *LogicAnd {
	if !l.GetType().IsBool() || !r.GetType().IsBool() {
		panic("unreachable")
	}
	inst := &LogicAnd{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self LogicAnd) inst() {}
func (self LogicAnd) GetType() Type {
	return self.Left.GetType()
}

// LogicOr 逻辑或
type LogicOr struct {
	Left, Right Value
}

func (self *Block) NewLogicOr(l, r Value) *LogicOr {
	if !l.GetType().IsBool() || !r.GetType().IsBool() {
		panic("unreachable")
	}
	inst := &LogicOr{
		Left:  l,
		Right: r,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self LogicOr) inst() {}
func (self LogicOr) GetType() Type {
	return self.Left.GetType()
}

// Call 调用
type Call struct {
	Func Value
	Args []Value
}

func (self *Block) NewCall(f Value, arg ...Value) *Call {
	ft := f.GetType()
	if !ft.IsFunc() {
		panic("unreachable")
	}
	params := ft.GetFuncParams()
	if (!ft.GetFuncVarArg() && len(params) != len(arg)) ||
		(ft.GetFuncVarArg() && len(params) > len(arg)) {
		panic("unreachable")
	}
	for i, p := range params {
		if !p.Equal(arg[i].GetType()) {
			panic("unreachable")
		}
	}

	inst := &Call{
		Func: f,
		Args: arg,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self Call) inst() {}
func (self Call) GetType() Type {
	return self.Func.GetType().GetFuncRet()
}

// WrapUnion 装进联合
type WrapUnion struct {
	To    Type
	Value Value
}

func (self *Block) NewWrapUnion(t Type, v Value) *WrapUnion {
	if !t.IsUnion() {
		panic("unreachable")
	}
	var exist bool
	vt := v.GetType()
	for _, e := range t.GetUnionElems() {
		if vt.Equal(e) {
			exist = true
			break
		}
	}
	if !exist {
		panic("unreachable")
	}

	inst := &WrapUnion{
		To:    t,
		Value: v,
	}
	self.Insts.PushBack(inst)
	return inst
}
func (self WrapUnion) inst() {}
func (self WrapUnion) GetType() Type {
	return self.To
}
