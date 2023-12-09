package mir

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"
)

// Stmt 语句
type Stmt interface {
	Define()string
	setIndex(i uint)
}

// AllocFromStack 从栈上分配内存
type AllocFromStack struct {
	i uint
	t Type
}

func (self *Builder) BuildAllocFromStack(t Type)*AllocFromStack{
	stmt := &AllocFromStack{t: t}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *AllocFromStack) Define()string{
	return fmt.Sprintf("%s %s = alloc %s from stack", self.Type(), self.Name(), self.t)
}

func (self *AllocFromStack) setIndex(i uint){
	self.i = i
}

func (self *AllocFromStack) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *AllocFromStack) Type()Type{
	return self.t.Context().NewPtrType(self.t)
}

func (self *AllocFromStack) ElemType()Type{
	return self.t
}

// AllocFromHeap 从堆上分配内存
type AllocFromHeap struct {
	i uint
	t Type
}

func (self *Builder) BuildAllocFromHeap(t Type)*AllocFromHeap{
	stmt := &AllocFromHeap{t: t}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *AllocFromHeap) Define()string{
	return fmt.Sprintf("%s %s = alloc %s from heap", self.Type(), self.Name(), self.t)
}

func (self *AllocFromHeap) setIndex(i uint){
	self.i = i
}

func (self *AllocFromHeap) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *AllocFromHeap) Type()Type{
	return self.t.Context().NewPtrType(self.t)
}

func (self *AllocFromHeap) ElemType()Type{
	return self.t
}

// Store 赋值
type Store struct {
	i uint
	from, to Value
}

func (self *Builder) BuildStore(from, to Value)*Store{
	if !from.Type().Equal(to.Type().(PtrType).Elem()){
		panic("unreachable")
	}
	stmt := &Store{
		from: from,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Store) Define()string{
	return fmt.Sprintf("store %s to %s", self.from.Name(), self.to.Name())
}

func (self *Store) setIndex(i uint){
	self.i = i
}

// Load 载入
type Load struct {
	i uint
	from Value
}

func (self *Builder) BuildLoad(from Value)*Load{
	if !stlbasic.Is[PtrType](from.Type()){
		panic("unreachable")
	}
	stmt := &Load{from: from}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Load) Define()string{
	return fmt.Sprintf("%s %s = load %s", self.Type(), self.Name(), self.from.Name())
}

func (self *Load) setIndex(i uint){
	self.i = i
}

func (self *Load) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Load) Type()Type{
	return self.from.Type().(PtrType).Elem()
}

// And 且
type And struct {
	i uint
	l, r Value
}

func (self *Builder) BuildAnd(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), new(big.Int).And(lc.IntValue(), rc.IntValue()))
		}
	}
	stmt := &And{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *And) Define()string{
	return fmt.Sprintf("%s %s = %s and %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *And) setIndex(i uint){
	self.i = i
}

func (self *And) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *And) Type()Type{
	return self.l.Type()
}

// Or 或
type Or struct {
	i uint
	l, r Value
}

func (self *Builder) BuildOr(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), new(big.Int).Or(lc.IntValue(), rc.IntValue()))
		}
	}
	stmt := &Or{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Or) Define()string{
	return fmt.Sprintf("%s %s = %s or %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Or) setIndex(i uint){
	self.i = i
}

func (self *Or) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Or) Type()Type{
	return self.l.Type()
}

// Xor 异或
type Xor struct {
	i uint
	l, r Value
}

func (self *Builder) BuildXor(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), new(big.Int).Xor(lc.IntValue(), rc.IntValue()))
		}
	}
	stmt := &Xor{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Xor) Define()string{
	return fmt.Sprintf("%s %s = %s xor %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Xor) setIndex(i uint){
	self.i = i
}

func (self *Xor) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Xor) Type()Type{
	return self.l.Type()
}

// Shl 左移
type Shl struct {
	i uint
	l, r Value
}

func (self *Builder) BuildShl(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), new(big.Int).Lsh(lc.IntValue(), uint(rc.IntValue().Uint64())))
		}
	}
	stmt := &Shl{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Shl) Define()string{
	return fmt.Sprintf("%s %s = %s shl %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Shl) setIndex(i uint){
	self.i = i
}

func (self *Shl) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Shl) Type()Type{
	return self.l.Type()
}

// Shr 右移
type Shr struct {
	i uint
	l, r Value
}

func (self *Builder) BuildShr(l, r Value)Value{
	if !stlbasic.Is[IntType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Int); ok{
		if rc, ok := r.(Int); ok{
			return NewInt(l.Type().(IntType), new(big.Int).Rsh(lc.IntValue(), uint(rc.IntValue().Uint64())))
		}
	}
	stmt := &Shr{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Shr) Define()string{
	return fmt.Sprintf("%s %s = %s shr %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Shr) setIndex(i uint){
	self.i = i
}

func (self *Shr) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Shr) Type()Type{
	return self.l.Type()
}

// Not 非

type Not struct {
	i uint
	v Value
}

func (self *Builder) BuildNot(v Value)Value{
	if !stlbasic.Is[IntType](v.Type()){
		panic("unreachable")
	}
	if vc, ok := v.(Int); ok{
		return NewInt(v.Type().(IntType), new(big.Int).Not(vc.IntValue()))
	}
	stmt := &Not{v: v}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Not) Define()string{
	return fmt.Sprintf("%s %s = not %s", self.Type(), self.Name(), self.v.Name())
}

func (self *Not) setIndex(i uint){
	self.i = i
}

func (self *Not) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Not) Type()Type{
	return self.v.Type()
}

// Add 加
type Add struct {
	i uint
	l, r Value
}

func (self *Builder) BuildAdd(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), new(big.Float).Add(lc.FloatValue(), rc.FloatValue()))
		}
	}
	stmt := &Add{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Add) Define()string{
	return fmt.Sprintf("%s %s = %s add %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Add) setIndex(i uint){
	self.i = i
}

func (self *Add) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Add) Type()Type{
	return self.l.Type()
}

// Sub 减
type Sub struct {
	i uint
	l, r Value
}

func (self *Builder) BuildSub(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), new(big.Float).Sub(lc.FloatValue(), rc.FloatValue()))
		}
	}
	stmt := &Sub{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Sub) Define()string{
	return fmt.Sprintf("%s %s = %s sub %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Sub) setIndex(i uint){
	self.i = i
}

func (self *Sub) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Sub) Type()Type{
	return self.l.Type()
}

// Mul 乘
type Mul struct {
	i uint
	l, r Value
}

func (self *Builder) BuildMul(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), new(big.Float).Mul(lc.FloatValue(), rc.FloatValue()))
		}
	}
	stmt := &Mul{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Mul) Define()string{
	return fmt.Sprintf("%s %s = %s mul %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Mul) setIndex(i uint){
	self.i = i
}

func (self *Mul) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Mul) Type()Type{
	return self.l.Type()
}

// Div 除
type Div struct {
	i uint
	l, r Value
}

func (self *Builder) BuildDiv(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			return NewNumber(l.Type().(NumberType), new(big.Float).Quo(lc.FloatValue(), rc.FloatValue()))
		}
	}
	stmt := &Div{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Div) Define()string{
	return fmt.Sprintf("%s %s = %s div %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Div) setIndex(i uint){
	self.i = i
}

func (self *Div) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Div) Type()Type{
	return self.l.Type()
}

// Rem 取余
type Rem struct {
	i uint
	l, r Value
}

func (self *Builder) BuildRem(l, r Value)Value{
	if !stlbasic.Is[NumberType](l.Type()) || !l.Type().Equal(r.Type()){
		panic("unreachable")
	}
	if lc, ok := l.(Number); ok{
		if rc, ok := r.(Number); ok{
			lcc, _ := lc.FloatValue().Float64()
			rcc, _ := rc.FloatValue().Float64()
			return NewNumber(l.Type().(NumberType), new(big.Float).SetFloat64(math.Mod(lcc, rcc)))
		}
	}
	stmt := &Rem{
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Rem) Define()string{
	return fmt.Sprintf("%s %s = %s rem %s", self.Type(), self.Name(), self.l.Name(), self.r.Name())
}

func (self *Rem) setIndex(i uint){
	self.i = i
}

func (self *Rem) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Rem) Type()Type{
	return self.l.Type()
}

// Neg 取反

type Neg struct {
	i uint
	v Value
}

func (self *Builder) BuildNeg(v Value)Value{
	if !stlbasic.Is[NumberType](v.Type()){
		panic("unreachable")
	}
	if vc, ok := v.(Number); ok{
		return NewNumber(v.Type().(NumberType), new(big.Float).Neg(vc.FloatValue()))
	}
	stmt := &Neg{v: v}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Neg) Define()string{
	return fmt.Sprintf("%s %s = neg %s", self.Type(), self.Name(), self.v.Name())
}

func (self *Neg) setIndex(i uint){
	self.i = i
}

func (self *Neg) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Neg) Type()Type{
	return self.v.Type()
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
			resInt := lc.FloatValue().Cmp(rc.FloatValue())
			var resBool bool
			switch kind {
			case CmpKindEQ:
				resBool = resInt == 0
			case CmpKindNE:
				resBool = resInt != 0
			case CmpKindLT:
				resBool = resInt < 0
			case CmpKindLE:
				resBool = resInt <= 0
			case CmpKindGT:
				resBool = resInt > 0
			case CmpKindGE:
				resBool = resInt >= 0
			default:
				panic("unreachable")
			}
			return Bool(l.Type().Context(), resBool)
		}
	}
	stmt := &Cmp{
		kind: kind,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *Cmp) Define()string{
	return fmt.Sprintf("%s %s = %s cmp %s %s", self.Type(), self.Name(), self.l.Name(), self.kind, self.r.Name())
}

func (self *Cmp) setIndex(i uint){
	self.i = i
}

func (self *Cmp) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Cmp) Type()Type{
	return self.l.Type().Context().Bool()
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
		kind: kind,
		l: l,
		r: r,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PtrEqual) Define()string{
	return fmt.Sprintf("%s %s = %s ptr %s %s", self.Type(), self.Name(), self.l.Name(), self.kind, self.r.Name())
}

func (self *PtrEqual) setIndex(i uint){
	self.i = i
}

func (self *PtrEqual) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *PtrEqual) Type()Type{
	return self.l.Type().Context().Bool()
}

// Call 调用

type Call struct {
	i uint
	f Value
	args []Value
}

func (self *Builder) BuildCall(f Value, arg ...Value)Value{
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
		f: f,
		args: arg,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
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
	return fmt.Sprintf("%%%d", self.i)
}

func (self *Call) Type()Type{
	return self.f.Type().(FuncType).Ret()
}

// NumberCovert 数字转换
type NumberCovert struct {
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
		switch tt := to.(type) {
		case SintType:
			intValue, _ := vc.FloatValue().Int(nil)
			return NewSint(tt, intValue)
		case UintType:
			intValue, _ := vc.FloatValue().Int(nil)
			return NewUint(tt, intValue)
		case FloatType:
			return NewFloat(tt, vc.FloatValue())
		default:
			panic("unreachable")
		}
	}
	stmt := &NumberCovert{
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *NumberCovert) Define()string{
	return fmt.Sprintf("%s %s = covert number %s to %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *NumberCovert) setIndex(i uint){
	self.i = i
}

func (self *NumberCovert) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *NumberCovert) Type()Type{
	return self.to
}

// PtrToUint 指针转uint
type PtrToUint struct {
	i uint
	v Value
	to UintType
}

func (self *Builder) BuildPtrToUint(v Value, to UintType)Value{
	if !stlbasic.Is[GenericPtrType](v.Type()){
		panic("unreachable")
	}
	stmt := &PtrToUint{
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *PtrToUint) Define()string{
	return fmt.Sprintf("%s %s = covert ptr %s to %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *PtrToUint) setIndex(i uint){
	self.i = i
}

func (self *PtrToUint) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *PtrToUint) Type()Type{
	return self.to
}

// UintToPtr uint转指针
type UintToPtr struct {
	i uint
	v Value
	to GenericPtrType
}

func (self *Builder) BuildUintToPtr(v Value, to GenericPtrType)Value{
	if !stlbasic.Is[Uint](v.Type()){
		panic("unreachable")
	}
	stmt := &UintToPtr{
		v: v,
		to: to,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *UintToPtr) Define()string{
	return fmt.Sprintf("%s %s = covert ptr %s from %s", self.Type(), self.Name(), self.v.Name(), self.to)
}

func (self *UintToPtr) setIndex(i uint){
	self.i = i
}

func (self *UintToPtr) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *UintToPtr) Type()Type{
	return self.to
}

// ArrayIndex 数组索引
type ArrayIndex struct {
	i uint
	v Value
	index Value
}

func (self *Builder) BuildArrayIndex(v, index Value)Value{
	if stlbasic.Is[ArrayType](v.Type()){
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[Array](v.Type().(PtrType).Elem()){
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
		v: v,
		index: index,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *ArrayIndex) Define()string{
	return fmt.Sprintf("%s %s = array %s index %s", self.Type(), self.Name(), self.v.Name(), self.index.Name())
}

func (self *ArrayIndex) setIndex(i uint){
	self.i = i
}

func (self *ArrayIndex) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *ArrayIndex) Type()Type{
	if stlbasic.Is[ArrayType](self.v.Type()){
		return self.v.Type().(ArrayType).Elem()
	}else{
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(ArrayType).Elem())
	}
}

// StructIndex 结构体索引
type StructIndex struct {
	i uint
	v Value
	index uint
}

func (self *Builder) BuildStructIndex(v Value, index uint)Value{
	var sizeLength uint
	if stlbasic.Is[StructType](v.Type()){
		sizeLength = uint(len(v.Type().(StructType).Elems()))
	}else if stlbasic.Is[PtrType](v.Type()) && stlbasic.Is[StructType](v.Type().(PtrType).Elem()){
		sizeLength = uint(len(v.Type().(PtrType).Elem().(StructType).Elems()))
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
		v: v,
		index: index,
	}
	self.cur.stmts.PushBack(stmt)
	return stmt
}

func (self *StructIndex) Define()string{
	return fmt.Sprintf("%s %s = struct %s index %d", self.Type(), self.Name(), self.v.Name(), self.index)
}

func (self *StructIndex) setIndex(i uint){
	self.i = i
}

func (self *StructIndex) Name()string{
	return fmt.Sprintf("%%%d", self.i)
}

func (self *StructIndex) Type()Type{
	if stlbasic.Is[StructType](self.v.Type()){
		return self.v.Type().(StructType).Elems()[self.index]
	}else{
		return self.v.Type().Context().NewPtrType(self.v.Type().(PtrType).Elem().(StructType).Elems()[self.index])
	}
}
