package hir

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/types"
	stlutil "github.com/kkkunny/stl/util"
)

type TypeKind uint8

const (
	TNone TypeKind = iota

	TBool

	TI8
	TU8
	TI16
	TU16
	TI32
	TU32
	TI64
	TU64
	TIsize
	TUsize

	TF32
	TF64

	TPtr
	TFunc
	TArray
	TTuple
	TStruct
	TEnum

	TTypedef
)

// Type 类型
type Type struct {
	Kind      TypeKind // 类型
	elems     []Type   // 子类型
	elemNames []string // 子类型名
	elemPubs  []bool   // 子类型公开性
	number    uint     // 数量
	typedef   *Typedef // 类型定义
	varArg    bool     // 是否是不定参数
}

func NewTypeNone() Type         { return Type{Kind: TNone} }
func NewTypeBool() Type         { return Type{Kind: TBool} }
func NewTypeI8() Type           { return Type{Kind: TI8} }
func NewTypeU8() Type           { return Type{Kind: TU8} }
func NewTypeI16() Type          { return Type{Kind: TI16} }
func NewTypeU16() Type          { return Type{Kind: TU16} }
func NewTypeI32() Type          { return Type{Kind: TI32} }
func NewTypeU32() Type          { return Type{Kind: TU32} }
func NewTypeI64() Type          { return Type{Kind: TI64} }
func NewTypeU64() Type          { return Type{Kind: TU64} }
func NewTypeIsize() Type        { return Type{Kind: TIsize} }
func NewTypeUsize() Type        { return Type{Kind: TUsize} }
func NewTypeF32() Type          { return Type{Kind: TF32} }
func NewTypeF64() Type          { return Type{Kind: TF64} }
func NewTypePtr(elem Type) Type { return Type{Kind: TPtr, elems: []Type{elem}} }
func NewTypeFunc(isVarArg bool, ret Type, params ...Type) Type {
	return Type{Kind: TFunc, elems: append([]Type{ret}, params...), varArg: isVarArg}
}
func NewTypeArray(size uint, elem Type) Type {
	return Type{Kind: TArray, elems: []Type{elem}, number: size}
}
func NewTypeTuple(elems ...Type) Type { return Type{Kind: TTuple, elems: elems} }
func NewTypeStruct(elems ...types.ThreePair[bool, string, Type]) Type {
	ps := make([]bool, len(elems))
	ns := make([]string, len(elems))
	ts := make([]Type, len(elems))
	for i, p := range elems {
		ps[i], ns[i], ts[i] = p.First, p.Second, p.Third
	}
	return Type{
		Kind:      TStruct,
		elems:     ts,
		elemNames: ns,
		elemPubs:  ps,
	}
}
func NewTypeEnum(elems ...types.ThreePair[bool, string, *Type]) Type {
	ps := make([]bool, len(elems))
	ns := make([]string, len(elems))
	ts := make([]Type, len(elems))
	for i, p := range elems {
		ps[i], ns[i] = p.First, p.Second
		if p.Third == nil {
			ts[i] = NewTypeNone()
		} else {
			ts[i] = *p.Third
		}
	}
	return Type{
		Kind:      TEnum,
		elems:     ts,
		elemNames: ns,
		elemPubs:  ps,
	}
}
func NewTypeTypedef(def *Typedef) Type { return Type{Kind: TTypedef, typedef: def} }

func (self Type) IsNone() bool { return self.Kind == TNone }
func (self Type) IsBool() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsBool()
	}
	return self.Kind == TBool
}
func (self Type) IsI8() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsI8()
	}
	return self.Kind == TI8
}
func (self Type) IsU8() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsU8()
	}
	return self.Kind == TU8
}
func (self Type) IsI16() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsI16()
	}
	return self.Kind == TI16
}
func (self Type) IsU16() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsU16()
	}
	return self.Kind == TU16
}
func (self Type) IsI32() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsI32()
	}
	return self.Kind == TI32
}
func (self Type) IsU32() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsU32()
	}
	return self.Kind == TU32
}
func (self Type) IsI64() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsI64()
	}
	return self.Kind == TI64
}
func (self Type) IsU64() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsU64()
	}
	return self.Kind == TU64
}
func (self Type) IsIsize() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsIsize()
	}
	return self.Kind == TIsize
}
func (self Type) IsUsize() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsUsize()
	}
	return self.Kind == TUsize
}
func (self Type) IsF32() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsF32()
	}
	return self.Kind == TF32
}
func (self Type) IsF64() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsF64()
	}
	return self.Kind == TF64
}
func (self Type) IsPtr() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsPtr()
	}
	return self.Kind == TPtr
}
func (self Type) IsFunc() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsFunc()
	}
	return self.Kind == TFunc
}
func (self Type) IsArray() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsArray()
	}
	return self.Kind == TArray
}
func (self Type) IsTuple() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsTuple()
	}
	return self.Kind == TTuple
}
func (self Type) IsStruct() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsStruct()
	}
	return self.Kind == TStruct
}
func (self Type) IsEnum() bool {
	if self.IsTypedef() {
		return self.GetTypedef().Target.IsEnum()
	}
	return self.Kind == TEnum
}
func (self Type) IsTypedef() bool { return self.Kind == TTypedef }

func (self Type) IsSint() bool {
	return self.IsI8() || self.IsI16() || self.IsI32() || self.IsI64() || self.IsIsize()
}
func (self Type) IsUint() bool {
	return self.IsU8() || self.IsU16() || self.IsU32() || self.IsU64() || self.IsUsize()
}
func (self Type) IsInt() bool    { return self.IsSint() || self.IsUint() }
func (self Type) IsFloat() bool  { return self.IsF32() || self.IsF64() }
func (self Type) IsNumber() bool { return self.IsInt() || self.IsFloat() }

func (self Type) String() string {
	switch self.Kind {
	case TBool:
		return "bool"
	case TI8:
		return "i8"
	case TU8:
		return "u8"
	case TI16:
		return "i16"
	case TU16:
		return "u16"
	case TI32:
		return "i32"
	case TU32:
		return "u32"
	case TI64:
		return "i64"
	case TU64:
		return "u64"
	case TIsize:
		return "isize"
	case TUsize:
		return "usize"
	case TF32:
		return "f32"
	case TF64:
		return "f64"
	case TPtr:
		return "*" + self.GetPtr().String()
	case TFunc:
		var buf strings.Builder
		buf.WriteString("func(")
		params := self.GetFuncParams()
		for i, p := range params {
			buf.WriteString(p.String())
			if i < len(params)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte(')')
		ret := self.GetFuncRet()
		if !ret.IsNone() {
			buf.WriteString(ret.String())
		}
		return buf.String()
	case TArray:
		return fmt.Sprintf("[%d]%s", self.GetArraySize(), self.GetArrayElem().String())
	case TTuple:
		var buf strings.Builder
		buf.WriteByte('(')
		elems := self.GetTupleElems()
		for i, e := range elems {
			buf.WriteString(e.String())
			if i < len(elems)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte(')')
		return buf.String()
	case TStruct:
		var buf strings.Builder
		buf.WriteByte('{')
		fields := self.GetStructFields()
		for i, f := range fields {
			buf.WriteString(f.Second)
			buf.WriteString(": ")
			buf.WriteString(f.Third.String())
			if i < len(fields)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte('}')
		return buf.String()
	case TEnum:
		var buf strings.Builder
		buf.WriteString("enum{")
		fields := self.GetEnumFields()
		for i, f := range fields {
			buf.WriteString(f.Second)
			if f.Third != nil {
				buf.WriteString(": ")
				buf.WriteString(f.Third.String())
			}
			if i < len(fields)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte('}')
		return buf.String()
	case TTypedef:
		def := self.GetTypedef()
		return fmt.Sprintf("%s.%s", def.Pkg, def.Name)
	default:
		panic("unreachable")
	}
}

// GetPtr 获取指针指向
func (self Type) GetPtr() Type {
	if !self.IsPtr() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetPtr()
	}
	return self.elems[0]
}

// GetFuncRet 获取函数返回值
func (self Type) GetFuncRet() Type {
	if !self.IsFunc() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetFuncRet()
	}
	return self.elems[0]
}

// GetFuncParams 获取函数参数
func (self Type) GetFuncParams() []Type {
	if !self.IsFunc() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetFuncParams()
	}
	return self.elems[1:]
}

// GetFuncVarArg 获取函数是否是不定参数
func (self Type) GetFuncVarArg() bool {
	return self.varArg
}

// GetArraySize 获取数组大小
func (self Type) GetArraySize() uint {
	if !self.IsArray() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetArraySize()
	}
	return self.number
}

// GetArrayElem 获取数组元素
func (self Type) GetArrayElem() Type {
	if !self.IsArray() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetArrayElem()
	}
	return self.elems[0]
}

// GetTupleElems 获取元组元素
func (self Type) GetTupleElems() []Type {
	if !self.IsTuple() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetTupleElems()
	}
	return self.elems
}

// GetStructFields 获取结构体字段
func (self Type) GetStructFields() []types.ThreePair[bool, string, Type] {
	if !self.IsStruct() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetStructFields()
	}
	pairs := make([]types.ThreePair[bool, string, Type], len(self.elems))
	for i, t := range self.elems {
		pairs[i] = types.NewThreePair(self.elemPubs[i], self.elemNames[i], t)
	}
	return pairs
}

// GetEnumFields 获取枚举字段
func (self Type) GetEnumFields() []types.ThreePair[bool, string, *Type] {
	if !self.IsEnum() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetEnumFields()
	}
	pairs := make([]types.ThreePair[bool, string, *Type], len(self.elems))
	for i, t := range self.elems {
		elemType := t
		pairs[i] = types.NewThreePair(
			self.elemPubs[i],
			self.elemNames[i],
			stlutil.Ternary[*Type](elemType.IsNone(), nil, &elemType),
		)
	}
	return pairs
}

// GetEnumFieldByName 获取指定枚举字段
func (self Type) GetEnumFieldByName(name string) (types.ThreePair[bool, string, *Type], bool) {
	if !self.IsEnum() {
		panic("unreachable")
	}
	if self.IsTypedef() {
		return self.GetTypedef().Target.GetEnumFieldByName(name)
	}
	for i, t := range self.elems {
		if self.elemNames[i] == name {
			elemType := t
			return types.NewThreePair(
				self.elemPubs[i],
				name,
				stlutil.Ternary[*Type](elemType.IsNone(), nil, &elemType),
			), true
		}
	}
	return types.ThreePair[bool, string, *Type]{}, false
}

// GetTypedef 获取定义类型
func (self Type) GetTypedef() *Typedef {
	if !self.IsTypedef() {
		panic("unreachable")
	}
	return self.typedef
}

// GetDeepTypedefTarget 获取深度定义类型目标
func (self Type) GetDeepTypedefTarget() Type {
	if !self.IsTypedef() {
		panic("unreachable")
	}

	cursor := self
	for cursor.IsTypedef() {
		cursor = cursor.GetTypedef().Target
	}
	return cursor
}

// Equal 比较
func (self Type) Equal(dst Type) bool {
	if self.Kind != dst.Kind {
		return false
	}

	switch self.Kind {
	case TNone, TBool, TI8, TU8, TI16, TU16, TI32, TU32, TI64, TU64, TIsize, TUsize, TF32, TF64:
		return true
	case TPtr:
		return self.GetPtr().Equal(dst.GetPtr())
	case TFunc:
		if !self.GetFuncRet().Equal(dst.GetFuncRet()) {
			return false
		}
		params1, params2 := self.GetFuncParams(), dst.GetFuncParams()
		if len(params1) != len(params2) {
			return false
		}
		for i, p := range params1 {
			if !p.Equal(params2[i]) {
				return false
			}
		}
		return true
	case TArray:
		return self.GetArraySize() == dst.GetArraySize() && self.GetArrayElem().Equal(dst.GetArrayElem())
	case TTuple:
		elems1, elems2 := self.GetTupleElems(), dst.GetTupleElems()
		if len(elems1) != len(elems2) {
			return false
		}
		for i, e := range elems1 {
			if !e.Equal(elems2[i]) {
				return false
			}
		}
		return true
	case TStruct:
		fields1, fields2 := self.GetStructFields(), dst.GetStructFields()
		if len(fields1) != len(fields2) {
			return false
		}
		for i, f := range fields1 {
			if f.Second != fields2[i].Second || !f.Third.Equal(fields2[i].Third) {
				return false
			}
		}
		return true
	case TEnum:
		fields1, fields2 := self.GetEnumFields(), dst.GetEnumFields()
		if len(fields1) != len(fields2) {
			return false
		}
		for i, f := range fields1 {
			if f.Second != fields2[i].Second {
				return false
			}
			if f.Third == nil && fields2[i].Third == nil {
				continue
			} else if f.Third != nil && fields2[i].Third != nil {
				if !f.Third.Equal(*fields2[i].Third) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	case TTypedef:
		def1, def2 := self.GetTypedef(), dst.GetTypedef()
		return def1.Pkg.Equal(def2.Pkg) && def1.Name == def2.Name
	default:
		panic("unreachable")
	}
}

// Like 近似
func (self Type) Like(dst Type) bool {
	for self.IsTypedef() && self.GetTypedef().Target.IsTypedef() {
		self = self.GetTypedef().Target
	}
	for dst.IsTypedef() && dst.GetTypedef().Target.IsTypedef() {
		dst = self.GetTypedef().Target
	}
	if self.IsTypedef() && dst.IsTypedef() && self.Equal(dst) {
		return true
	}

	if self.IsTypedef() {
		self = self.GetTypedef().Target
	}
	if dst.IsTypedef() {
		dst = dst.GetTypedef().Target
	}

	if self.Kind != dst.Kind {
		return false
	}

	switch self.Kind {
	case TNone, TBool, TI8, TU8, TI16, TU16, TI32, TU32, TI64, TU64, TIsize, TUsize, TF32, TF64:
		return true
	case TPtr:
		return self.GetPtr().Like(dst.GetPtr())
	case TFunc:
		if !self.GetFuncRet().Like(dst.GetFuncRet()) {
			return false
		}
		params1, params2 := self.GetFuncParams(), dst.GetFuncParams()
		if len(params1) != len(params2) {
			return false
		}
		for i, p := range params1 {
			if !p.Like(params2[i]) {
				return false
			}
		}
		return true
	case TArray:
		return self.GetArraySize() == dst.GetArraySize() && self.GetArrayElem().Like(dst.GetArrayElem())
	case TTuple:
		elems1, elems2 := self.GetTupleElems(), dst.GetTupleElems()
		if len(elems1) != len(elems2) {
			return false
		}
		for i, e := range elems1 {
			if !e.Like(elems2[i]) {
				return false
			}
		}
		return true
	case TStruct:
		fields1, fields2 := self.GetStructFields(), dst.GetStructFields()
		if len(fields1) != len(fields2) {
			return false
		}
		for i, f := range fields1 {
			if f.Second != fields2[i].Second || !f.Third.Like(fields2[i].Third) {
				return false
			}
		}
		return true
	case TEnum:
		fields1, fields2 := self.GetEnumFields(), dst.GetEnumFields()
		if len(fields1) != len(fields2) {
			return false
		}
		for i, f := range fields1 {
			if f.Second != fields2[i].Second {
				return false
			}
			if f.Third == nil && fields2[i].Third == nil {
				continue
			} else if f.Third != nil && fields2[i].Third != nil {
				if !f.Third.Like(*fields2[i].Third) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	default:
		panic("unreachable")
	}
}
