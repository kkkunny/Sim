package mir

import (
	"fmt"
	"math"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	stlbits "github.com/kkkunny/stl/math/bits"
	"github.com/samber/lo"
)

// Stmt 语句
type Stmt interface {
	Define()string
	Belong()*Block
}

type StmtValue interface {
	Stmt
	Value
	setIndex(i uint)
}

// AllocFromStack 从栈上分配内存
type AllocFromStack struct {
	b *Block
	i uint
	t Type
}

func (self *Builder) BuildAllocFromStack(t Type)*AllocFromStack{
	stmt := &AllocFromStack{
		b: self.cur,
		t: t,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *AllocFromStack) Belong()*Block{
	return self.b
}

func (self *AllocFromStack) Define()string{
	return fmt.Sprintf("%s %s = alloc %s from stack", self.Type(), self.Name(), self.t)
}

func (self *AllocFromStack) setIndex(i uint){
	self.i = i
}

func (self *AllocFromStack) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *AllocFromStack) Type()Type{
	return self.t.Context().NewPtrType(self.t)
}

func (self *AllocFromStack) ElemType()Type{
	return self.t
}

// AllocFromHeap 从堆上分配内存
type AllocFromHeap struct {
	b *Block
	i uint
	t Type
}

func (self *Builder) BuildAllocFromHeap(t Type)*AllocFromHeap{
	stmt := &AllocFromHeap{
		b: self.cur,
		t: t,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *AllocFromHeap) Belong()*Block{
	return self.b
}

func (self *AllocFromHeap) Define()string{
	return fmt.Sprintf("%s %s = alloc %s from heap", self.Type(), self.Name(), self.t)
}

func (self *AllocFromHeap) setIndex(i uint){
	self.i = i
}

func (self *AllocFromHeap) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *AllocFromHeap) Type()Type{
	return self.t.Context().NewPtrType(self.t)
}

func (self *AllocFromHeap) ElemType()Type{
	return self.t
}

// Store 赋值
type Store struct {
	b *Block
	from, to Value
}

func (self *Builder) BuildStore(from, to Value)Stmt{
	if !from.Type().Equal(to.Type().(PtrType).Elem()){
		panic("unreachable")
	}
	stmt := &Store{
		b: self.cur,
		from: from,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Store) Belong()*Block{
	return self.b
}

func (self *Store) Define()string{
	return fmt.Sprintf("store %s to %s", self.from.Name(), self.to.Name())
}

func (self *Store) From()Value{
	return self.from
}

func (self *Store) To()Value{
	return self.to
}

// Load 载入
type Load struct {
	b *Block
	i uint
	from Value
}

func (self *Builder) BuildLoad(from Value)StmtValue{
	if !stlbasic.Is[PtrType](from.Type()){
		panic("unreachable")
	}
	stmt := &Load{
		b: self.cur,
		from: from,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Load) Belong()*Block{
	return self.b
}

func (self *Load) Define()string{
	return fmt.Sprintf("%s %s = load %s", self.Type(), self.Name(), self.from.Name())
}

func (self *Load) setIndex(i uint){
	self.i = i
}

func (self *Load) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Load) Type()Type{
	return self.from.Type().(PtrType).Elem()
}

func (self *Load) From()Value{
	return self.from
}

// And 且
type And struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildAnd(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), lc.IntValue()&rc.IntValue())
		}
	}
	stmt := &And{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *And) Belong()*Block{
	return self.b
}

func (self *And) Define()string{
	return fmt.Sprintf("%s %s = %s and %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *And) setIndex(i uint){
	self.i = i
}

func (self *And) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *And) Type()Type{
	return self.l.Type()
}

func (self *And) Left()Value{
	return self.l
}

func (self *And) Right()Value{
	return self.r
}

// Or 或
type Or struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildOr(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), lc.IntValue()|rc.IntValue())
		}
	}
	stmt := &Or{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Or) Belong()*Block{
	return self.b
}

func (self *Or) Define()string{
	return fmt.Sprintf("%s %s = %s or %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Or) setIndex(i uint){
	self.i = i
}

func (self *Or) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Or) Type()Type{
	return self.l.Type()
}

func (self *Or) Left()Value{
	return self.l
}

func (self *Or) Right()Value{
	return self.r
}

// Xor 异或
type Xor struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildXor(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), lc.IntValue()^rc.IntValue())
		}
	}
	stmt := &Xor{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Xor) Belong()*Block{
	return self.b
}

func (self *Xor) Define()string{
	return fmt.Sprintf("%s %s = %s xor %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Xor) setIndex(i uint){
	self.i = i
}

func (self *Xor) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Xor) Type()Type{
	return self.l.Type()
}

func (self *Xor) Left()Value{
	return self.l
}

func (self *Xor) Right()Value{
	return self.r
}

// Shl 左移
type Shl struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildShl(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), lc.IntValue()<<rc.IntValue())
		}
	}
	stmt := &Shl{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Shl) Belong()*Block{
	return self.b
}

func (self *Shl) Define()string{
	return fmt.Sprintf("%s %s = %s shl %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Shl) setIndex(i uint){
	self.i = i
}

func (self *Shl) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Shl) Type()Type{
	return self.l.Type()
}

func (self *Shl) Left()Value{
	return self.l
}

func (self *Shl) Right()Value{
	return self.r
}

// Shr 右移
type Shr struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildShr(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), lc.IntValue()>>rc.IntValue())
		}
	}
	stmt := &Shr{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Shr) Belong()*Block{
	return self.b
}

func (self *Shr) Define()string{
	return fmt.Sprintf("%s %s = %s shr %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Shr) setIndex(i uint){
	self.i = i
}

func (self *Shr) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Shr) Type()Type{
	return self.l.Type()
}

func (self *Shr) Left()Value{
	return self.l
}

func (self *Shr) Right()Value{
	return self.r
}

// Not 非

type Not struct {
	b *Block
	i uint
	v Value
}

func (self *Builder) BuildNot(v Value)Value{
	if !stlbasic.Is[IntType](v.Type()){
		panic("unreachable")
	}
	if vc, ok := v.(Int); ok{
		if stlbasic.Is[*Sint](vc){
			return NewSint(v.Type().(SintType), stlbits.NotWithLength(vc.IntValue(), uint64(vc.Type().Size())))
		}else{
			return NewUint(v.Type().(UintType), stlbits.NotWithLength(uint64(vc.IntValue()), uint64(vc.Type().Size())))
		}
	}
	stmt := &Not{
		b: self.cur,
		v: v,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Not) Belong()*Block{
	return self.b
}

func (self *Not) Define()string{
	return fmt.Sprintf("%s %s = not %s", self.Type(), self.Name(), self.v.Name())
}

func (self *Not) setIndex(i uint){
	self.i = i
}

func (self *Not) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Not) Type()Type{
	return self.v.Type()
}

func (self *Not) Value()Value{
	return self.v
}

// Add 加
type Add struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildAdd(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), lc.FloatValue()+rc.FloatValue())
		}
	}
	stmt := &Add{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Add) Belong()*Block{
	return self.b
}

func (self *Add) Define()string{
	return fmt.Sprintf("%s %s = %s add %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Add) setIndex(i uint){
	self.i = i
}

func (self *Add) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Add) Type()Type{
	return self.l.Type()
}

func (self *Add) Left()Value{
	return self.l
}

func (self *Add) Right()Value{
	return self.r
}

// Sub 减
type Sub struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildSub(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), lc.FloatValue()-rc.FloatValue())
		}
	}
	stmt := &Sub{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Sub) Belong()*Block{
	return self.b
}

func (self *Sub) Define()string{
	return fmt.Sprintf("%s %s = %s sub %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Sub) setIndex(i uint){
	self.i = i
}

func (self *Sub) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Sub) Type()Type{
	return self.l.Type()
}

func (self *Sub) Left()Value{
	return self.l
}

func (self *Sub) Right()Value{
	return self.r
}

// Mul 乘
type Mul struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildMul(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), lc.FloatValue()*rc.FloatValue())
		}
	}
	stmt := &Mul{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Mul) Belong()*Block{
	return self.b
}

func (self *Mul) Define()string{
	return fmt.Sprintf("%s %s = %s mul %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Mul) setIndex(i uint){
	self.i = i
}

func (self *Mul) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Mul) Type()Type{
	return self.l.Type()
}

func (self *Mul) Left()Value{
	return self.l
}

func (self *Mul) Right()Value{
	return self.r
}

// Div 除
type Div struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildDiv(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), lc.FloatValue()/rc.FloatValue())
		}
	}
	stmt := &Div{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Div) Belong()*Block{
	return self.b
}

func (self *Div) Define()string{
	return fmt.Sprintf("%s %s = %s div %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Div) setIndex(i uint){
	self.i = i
}

func (self *Div) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Div) Type()Type{
	return self.l.Type()
}

func (self *Div) Left()Value{
	return self.l
}

func (self *Div) Right()Value{
	return self.r
}

// Rem 取余
type Rem struct {
	b *Block
	i uint
	l, r Value
}

func (self *Builder) BuildRem(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), math.Mod(lc.FloatValue(), rc.FloatValue()))
		}
	}
	stmt := &Rem{
		b: self.cur,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Rem) Belong()*Block{
	return self.b
}

func (self *Rem) Define()string{
	return fmt.Sprintf("%s %s = %s rem %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Rem) setIndex(i uint){
	self.i = i
}

func (self *Rem) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Rem) Type()Type{
	return self.l.Type()
}

func (self *Rem) Left()Value{
	return self.l
}

func (self *Rem) Right()Value{
	return self.r
}

func (self *Builder) BuildNeg(v Value)Value{
	if !stlbasic.Is[NumberType](v.Type()){
		panic("unreachable")
	}
	return self.BuildSub(NewZero(v.Type()), v)
}

type CmpKind uint8

const (
	CmpKindEQ CmpKind = iota
	CmpKindNE
	CmpKindLT
	CmpKindLE
	CmpKindGT
	CmpKindGE
)

func (self CmpKind) String()string{
	switch self {
	case CmpKindEQ:
		return "eq"
	case CmpKindNE:
		return "ne"
	case CmpKindLT:
		return "lt"
	case CmpKindLE:
		return "le"
	case CmpKindGT:
		return "gt"
	case CmpKindGE:
		return "ge"
	default:
		panic("unreachable")
	}
}

// Cmp 比较
type Cmp struct {
	b *Block
	i uint
	kind CmpKind
	l, r Value
}

func (self *Builder) BuildCmp(kind CmpKind, l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			var resBool bool
			switch kind {
			case CmpKindEQ:
				resBool = lc.FloatValue() == rc.FloatValue()
			case CmpKindNE:
				resBool = lc.FloatValue() != rc.FloatValue()
			case CmpKindLT:
				resBool = lc.FloatValue() < rc.FloatValue()
			case CmpKindLE:
				resBool = lc.FloatValue() <= rc.FloatValue()
			case CmpKindGT:
				resBool = lc.FloatValue() > rc.FloatValue()
			case CmpKindGE:
				resBool = lc.FloatValue() >= rc.FloatValue()
			default:
				panic("unreachable")
			}
			return Bool(l.Type().Context(), resBool)
		}
	}
	stmt := &Cmp{
		b: self.cur,
		kind: kind,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Cmp) Belong()*Block{
	return self.b
}

func (self *Cmp) Define()string{
	return fmt.Sprintf("%s %s = %s cmp %s %s", self.Type(), self.Name(), self.l.Name(), self.kind, self.r.Name())
}

func (self *Cmp) setIndex(i uint){
	self.i = i
}

func (self *Cmp) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Cmp) Type()Type{
	return self.l.Type().Context().Bool()
}

func (self *Cmp) Kind()CmpKind{
	return self.kind
}

func (self *Cmp) Left()Value{
	return self.l
}

func (self *Cmp) Right()Value{
	return self.r
}

type PtrEqualKind uint8

const (
	PtrEqualKindEQ PtrEqualKind = iota
	PtrEqualKindNE
)

func (self PtrEqualKind) String()string{
	switch self {
	case PtrEqualKindEQ:
		return "eq"
	case PtrEqualKindNE:
		return "ne"
	default:
		panic("unreachable")
	}
}

// PtrEqual 指针比较

type PtrEqual struct {
	b *Block
	i uint
	kind PtrEqualKind
	l, r Value
}

func (self *Builder) BuildPtrEqual(kind PtrEqualKind, l, r Value)Value{
	if !stlbasic.Is[GenericPtrType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if l == r{
		return Bool(l.Type().Context(), kind==PtrEqualKindEQ)
	}
	stmt := &PtrEqual{
		b: self.cur,
		kind: kind,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PtrEqual) Belong()*Block{
	return self.b
}

func (self *PtrEqual) Define()string{
	return fmt.Sprintf("%s %s = %s ptr %s %s", self.Type(), self.Name(), self.l.Name(), self.kind, self.r.Name())
}

func (self *PtrEqual) setIndex(i uint){
	self.i = i
}

func (self *PtrEqual) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *PtrEqual) Type()Type{
	return self.l.Type().Context().Bool()
}

func (self *PtrEqual) Kind()PtrEqualKind{
	return self.kind
}

func (self *PtrEqual) Left()Value{
	return self.l
}

func (self *PtrEqual) Right()Value{
	return self.r
}

// Call 调用

type Call struct {
	b *Block
	i uint
	f Value
	args []Value
}

func (self *Builder) BuildCall(f Value, arg ...Value)StmtValue{
	params := f.Type().(FuncType).Params()
	if len(params) != len(arg){
		panic("unreachable")
	}
	for i, p := range params{
		if !p.Equal(arg[i].Type()){
			panic("unreachable")
		}
	}
	stmt := &Call{
		b: self.cur,
		f: f,
		args: arg,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Call) Belong()*Block{
	return self.b
}

func (self *Call) Define()string{
	args := lo.Map(self.args, func(item Value, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("%s %s = call %s(%s)", self.Type(), self.Name(), self.f.Name(), strings.Join(args, ","))
}

func (self *Call) setIndex(i uint){
	self.i = i
}

func (self *Call) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Call) Type()Type{
	return self.f.Type().(FuncType).Ret()
}

func (self *Call) Func()Value{
	return self.f
}

func (self *Call) Args()[]Value{
	return self.args
}

// NumberCovert 数字转换
type NumberCovert struct {
	b *Block
	i uint
	v Value
	to NumberType
}

func (self *Builder) BuildNumberCovert(v Value, to NumberType)Value{
	if !stlbasic.Is[NumberType](v.Type()){
		panic("unreachable")
	}
	if v.Type().Equal(to){
		return v
	}
	if vc, ok := v.(Number); ok{
		return NewNumber(to, vc.FloatValue())
	}
	stmt := &NumberCovert{
		b: self.cur,
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *NumberCovert) Belong()*Block{
	return self.b
}

func (self *NumberCovert) Define()string{
	return fmt.Sprintf("%s %s = covert Bits %s to %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *NumberCovert) setIndex(i uint){
	self.i = i
}

func (self *NumberCovert) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *NumberCovert) Type()Type{
	return self.to
}

func (self *NumberCovert) From()Value{
	return self.v
}

// PtrToUint 指针转uint
type PtrToUint struct {
	b *Block
	i uint
	v Value
	to UintType
}

func (self *Builder) BuildPtrToUint(v Value, to UintType)StmtValue{
	if !stlbasic.Is[GenericPtrType](v.Type()){
		panic("unreachable")
	}
	stmt := &PtrToUint{
		b: self.cur,
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PtrToUint) Belong()*Block{
	return self.b
}

func (self *PtrToUint) Define()string{
	return fmt.Sprintf("%s %s = covert ptr %s to %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *PtrToUint) setIndex(i uint){
	self.i = i
}

func (self *PtrToUint) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *PtrToUint) Type()Type{
	return self.to
}

func (self *PtrToUint) From()Value{
	return self.v
}

// UintToPtr uint转指针
type UintToPtr struct {
	b *Block
	i uint
	v Value
	to GenericPtrType
}

func (self *Builder) BuildUintToPtr(v Value, to GenericPtrType)StmtValue{
	if !stlbasic.Is[Uint](v.Type()){
		panic("unreachable")
	}
	stmt := &UintToPtr{
		b: self.cur,
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *UintToPtr) Belong()*Block{
	return self.b
}

func (self *UintToPtr) Define()string{
	return fmt.Sprintf("%s %s = covert ptr %s from %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *UintToPtr) setIndex(i uint){
	self.i = i
}

func (self *UintToPtr) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *UintToPtr) Type()Type{
	return self.to
}

func (self *UintToPtr) From()Value{
	return self.v
}

// PtrToPtr 指针转指针
type PtrToPtr struct {
	b *Block
	i uint
	v Value
	to GenericPtrType
}

func (self *Builder) BuildPtrToPtr(v Value, to GenericPtrType)Value{
	if !stlbasic.Is[GenericPtrType](v.Type()){
		panic("unreachable")
	}
	if v.Type().Equal(to){
		return v
	}
	stmt := &PtrToPtr{
		b: self.cur,
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PtrToPtr) Belong()*Block{
	return self.b
}

func (self *PtrToPtr) Define()string{
	return fmt.Sprintf("%s %s = covert ptr %s to %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *PtrToPtr) setIndex(i uint){
	self.i = i
}

func (self *PtrToPtr) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *PtrToPtr) Type()Type{
	return self.to
}

func (self *PtrToPtr) From()Value{
	return self.v
}

// ArrayIndex 数组索引
type ArrayIndex struct {
	b *Block
	i uint
	v Value
	index Value
}

func (self *Builder) BuildArrayIndex(v, index Value)Value{
	if stlbasic.Is[ArrayType](v.Type()){
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[ArrayType](v.Type().(PtrType).Elem()){
	}else{
		panic("unreachable")
	}
	if !stlbasic.Is[UintType](index.Type()){
		panic("unreachable")
	}
	if vc, ok := v.(Const); ok{
		if ic, ok := index.(Const); ok{
			return NewArrayIndex(vc, ic)
		}
	}
	stmt := &ArrayIndex{
		b: self.cur,
		v: v,
		index: index,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *ArrayIndex) Belong()*Block{
	return self.b
}

func (self *ArrayIndex) Define()string{
	return fmt.Sprintf("%s %s = array %s index %s", self.Type(), self.Name(), self.v.Name(), self.index.Name())
}

func (self *ArrayIndex) setIndex(i uint){
	self.i = i
}

func (self *ArrayIndex) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *ArrayIndex) Type()Type{
	if self.IsPtr(){
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(ArrayType).Elem())
	}
	return self.v.Type().(ArrayType).Elem()
}

func (self *ArrayIndex) IsPtr()bool{
	return stlbasic.Is[PtrType](self.v.Type())
}

func (self *ArrayIndex) Array()Value{
	return self.v
}

func (self *ArrayIndex) Index()Value{
	return self.index
}

// StructIndex 结构体索引
type StructIndex struct {
	b *Block
	i uint
	v Value
	index uint64
}

func (self *Builder) BuildStructIndex(v Value, index uint64)Value{
	var sizeLength uint64
	if stlbasic.Is[StructType](v.Type()){
		sizeLength = uint64(len(v.Type().(StructType).Elems()))
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[StructType](v.Type().(PtrType).Elem()){
		sizeLength = uint64(len(v.Type().(PtrType).Elem().(StructType).Elems()))
	}else{
		panic("unreachable")
	}
	if index >= sizeLength{
		panic("unreachable")
	}
	if vc, ok := v.(Const); ok{
		return NewStructIndex(vc, index)
	}
	stmt := &StructIndex{
		b: self.cur,
		v: v,
		index: index,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *StructIndex) Belong()*Block{
	return self.b
}

func (self *StructIndex) Define()string{
	return fmt.Sprintf("%s %s = struct %s index %d", self.Type(), self.Name(), self.v.Name(), self.index)
}

func (self *StructIndex) setIndex(i uint){
	self.i = i
}

func (self *StructIndex) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *StructIndex) Type()Type{
	if self.IsPtr(){
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(StructType).Elems()[self.index])
	}
	return self.v.Type().(StructType).Elems()[self.index]
}

func (self *StructIndex) IsPtr()bool{
	return stlbasic.Is[PtrType](self.v.Type())
}

func (self *StructIndex) Struct()Value{
	return self.v
}

func (self *StructIndex) Index()uint64{
	return self.index
}

type Terminating interface {
	Stmt
	terminate()
}

// Return 返回
type Return struct {
	b *Block
	v Value
}

func (self *Builder) BuildReturn(v ...Value)Terminating{
	var value Value
	if len(v) != 0{
		value = v[0]
	}
	if value == nil && self.cur.f.t.Ret().Equal(self.ctx.Void()){
	}else if value != nil && value.Type().Equal(self.cur.f.t.Ret()){
	}else{
		panic("unreachable")
	}
	stmt := &Return{
		b: self.cur,
		v: value,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Return) Belong()*Block{
	return self.b
}

func (self *Return) Define()string{
	if self.v == nil{
		return "ret"
	}
	return fmt.Sprintf("ret %s", self.v.Name())
}

func (self *Return) Value()(Value, bool){
	return self.v, self.v!=nil
}

func (*Return)terminate(){}

type Jump interface {
	Terminating
	Targets()[]*Block
}

// UnCondJump 无条件跳转
type UnCondJump struct {
	b *Block
	to *Block
}

func (self *Builder) BuildUnCondJump(to *Block)Jump {
	stmt := &UnCondJump{
		b: self.cur,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *UnCondJump) Belong()*Block{
	return self.b
}

func (self *UnCondJump) Define()string{
	return fmt.Sprintf("jump %s", self.to.Name())
}

func (self *UnCondJump) To()*Block{
	return self.to
}

func (*UnCondJump)terminate(){}

func (self *UnCondJump) Targets()[]*Block{
	return []*Block{self.to}
}

// CondJump 条件跳转
type CondJump struct {
	b *Block
	cond Value
	trueTo, falseTo *Block
}

func (self *Builder) BuildCondJump(cond Value, trueTo, falseTo *Block)Jump {
	if !cond.Type().Equal(self.ctx.Bool()){
		panic("unreachable")
	}
	if trueTo == falseTo{
		return self.BuildUnCondJump(trueTo)
	}
	if cc, ok := cond.(*Uint); ok{
		if cc.IsZero(){
			return self.BuildUnCondJump(falseTo)
		}else{
			return self.BuildUnCondJump(trueTo)
		}
	}
	stmt := &CondJump{
		b: self.cur,
		cond: cond,
		trueTo: trueTo,
		falseTo: falseTo,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *CondJump) Belong()*Block{
	return self.b
}

func (self *CondJump) Define()string{
	return fmt.Sprintf("if %s jump %s or %s", self.cond.Name(), self.trueTo.Name(), self.falseTo.Name())
}

func (self *CondJump) Cond()Value{
	return self.cond
}

func (self *CondJump) TrueBlock()*Block{
	return self.trueTo
}

func (self *CondJump) FalseBlock()*Block{
	return self.falseTo
}

func (*CondJump)terminate(){}

func (self *CondJump) Targets()[]*Block{
	return []*Block{self.trueTo, self.falseTo}
}

// Phi 跳转收拢
type Phi struct {
	b *Block
	i uint
	t Type
	froms []pair.Pair[*Block, Value]
}

func (self *Builder) BuildPhi(t Type, from ...pair.Pair[*Block, Value])*Phi {
	if self.cur.stmts.Len() != 0{
		panic("unreachable")
	}
	for _, f := range from{
		if !f.Second.Type().Equal(t){
			panic("unreachable")
		}
	}
	stmt := &Phi{
		b: self.cur,
		t: t,
		froms: from,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Phi) Belong()*Block{
	return self.b
}

func (self *Phi) Define()string{
	froms := lo.Map(self.froms, func(item pair.Pair[*Block, Value], _ int) string {
		return fmt.Sprintf("%s:%s", item.First.Name(), item.Second.Name())
	})
	return fmt.Sprintf("%s %s = phi [%s]", self.t, self.Name(), strings.Join(froms, ","))
}

func (self *Phi) setIndex(i uint){
	self.i = i
}

func (self *Phi) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *Phi) Type()Type{
	return self.t
}

func (self *Phi) AddFroms(from ...pair.Pair[*Block, Value]){
	for _, f := range from{
		if !f.Second.Type().Equal(self.t){
			panic("unreachable")
		}
	}
	self.froms = append(self.froms, from...)
}

func (self *Phi) Froms()[]pair.Pair[*Block, Value]{
	return self.froms
}

// PackArray 打包数组
type PackArray struct {
	b *Block
	i uint
	t ArrayType
	elems []Value
}

func (self *Builder) BuildPackArray(t ArrayType, elem ...Value)Value {
	if t.Length() != uint(len(elem)){
		panic("unreachable")
	}
	for _, e := range elem{
		if !e.Type().Equal(t.Elem()){
			panic("unreachable")
		}
	}
	allConst := true
	constElems := make([]Const, len(elem))
	for i, e := range elem{
		if c, ok := e.(Const); ok{
			constElems[i] = c
		}else{
			allConst = false
		}
	}
	if allConst{
		return NewArray(t, constElems...)
	}
	stmt := &PackArray{
		b: self.cur,
		t: t,
		elems: elem,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PackArray) Belong()*Block{
	return self.b
}

func (self *PackArray) Define()string{
	elems := lo.Map(self.elems, func(item Value, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("%s %s = unpack (%s)", self.t, self.Name(), strings.Join(elems, ","))
}

func (self *PackArray) setIndex(i uint){
	self.i = i
}

func (self *PackArray) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *PackArray) Type()Type{
	return self.t
}

func (self *PackArray) Elems()[]Value{
	return self.elems
}

// PackStruct 打包结构体
type PackStruct struct {
	b *Block
	i uint
	t StructType
	elems []Value
}

func (self *Builder) BuildPackStruct(t StructType, elem ...Value)Value {
	if len(t.Elems()) != len(elem){
		panic("unreachable")
	}
	for i, e := range elem{
		if !e.Type().Equal(t.Elems()[i]){
			panic("unreachable")
		}
	}
	allConst := true
	constElems := make([]Const, len(elem))
	for i, e := range elem{
		if c, ok := e.(Const); ok{
			constElems[i] = c
		}else{
			allConst = false
		}
	}
	if allConst{
		return NewStruct(t, constElems...)
	}
	stmt := &PackStruct{
		b: self.cur,
		t: t,
		elems: elem,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PackStruct) Belong()*Block{
	return self.b
}

func (self *PackStruct) Define()string{
	elems := lo.Map(self.elems, func(item Value, _ int) string {
		return item.Name()
	})
	return fmt.Sprintf("%s %s = unpack {%s}", self.t, self.Name(), strings.Join(elems, ","))
}

func (self *PackStruct) setIndex(i uint){
	self.i = i
}

func (self *PackStruct) Name()string{
	return fmt.Sprintf("%%%s_%d", self.b.Name(), self.i)
}

func (self *PackStruct) Type()Type{
	return self.t
}

func (self *PackStruct) Elems()[]Value{
	return self.elems
}
