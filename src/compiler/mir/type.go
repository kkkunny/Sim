package mir

import (
	"math"

	"github.com/kkkunny/Sim/src/compiler/utils"
)

type TypeKind uint8

const (
	TVoid TypeKind = iota
	TBool
	TSint
	TUint
	TFloat
	TPtr
	TFunc
	TArray
	TStruct
	TUnion
	TAlias
)

// Type 类型
type Type struct {
	Kind   TypeKind // 类型
	width  uint     // 宽度
	elems  []Type   // 子类型
	varArg bool     // 是否是不定参数
	alias  *Alias   // 别名
}

func NewTypeVoid() Type       { return Type{Kind: TVoid} }
func NewTypeBool() Type       { return Type{Kind: TBool} }
func NewTypeSint(w uint) Type { return Type{Kind: TSint, width: w} }
func NewTypeUint(w uint) Type { return Type{Kind: TUint, width: w} }
func NewTypeFloat(w uint) Type {
	if w == 0 || (w != 4 && w != 8) {
		panic("unreachable")
	}
	return Type{Kind: TFloat, width: w}
}
func NewTypePtr(elem Type) Type { return Type{Kind: TPtr, elems: []Type{elem}} }
func NewTypeFunc(isVarArg bool, ret Type, params ...Type) Type {
	return Type{Kind: TFunc, elems: append([]Type{ret}, params...), varArg: isVarArg}
}
func NewTypeArray(size uint, elem Type) Type {
	return Type{Kind: TArray, elems: []Type{elem}, width: size}
}
func NewTypeStruct(elem ...Type) Type { return Type{Kind: TStruct, elems: elem} }
func NewTypeUnion(elem ...Type) Type  { return Type{Kind: TUnion, elems: elem} }
func NewTypeAlias(a *Alias) Type      { return Type{Kind: TAlias, alias: a} }

func (self Type) IsVoid() bool { return self.Kind == TVoid }
func (self Type) IsBool() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsBool()
	}
	return self.Kind == TBool
}
func (self Type) IsSint() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsSint()
	}
	return self.Kind == TSint
}
func (self Type) IsUint() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsUint()
	}
	return self.Kind == TUint
}
func (self Type) IsFloat() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsFloat()
	}
	return self.Kind == TFloat
}
func (self Type) IsPtr() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsPtr()
	}
	return self.Kind == TPtr
}
func (self Type) IsFunc() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsFunc()
	}
	return self.Kind == TFunc
}
func (self Type) IsArray() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsArray()
	}
	return self.Kind == TArray
}
func (self Type) IsStruct() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsStruct()
	}
	return self.Kind == TStruct
}
func (self Type) IsUnion() bool {
	if self.IsAlias() {
		return self.GetAliasTarget().IsUnion()
	}
	return self.Kind == TUnion
}
func (self Type) IsAlias() bool { return self.Kind == TAlias }

func (self Type) IsInteger() bool { return self.IsSint() || self.IsUint() }
func (self Type) IsNumber() bool  { return self.IsInteger() || self.IsFloat() }

// GetWidth 获取宽度
func (self Type) GetWidth() uint {
	if !self.IsNumber() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetWidth()
	}
	return self.width
}

// GetPtr 获取指针指向
func (self Type) GetPtr() Type {
	if !self.IsPtr() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetPtr()
	}
	return self.elems[0]
}

// GetFuncRet 获取函数返回值
func (self Type) GetFuncRet() Type {
	if !self.IsFunc() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetFuncRet()
	}
	return self.elems[0]
}

// GetFuncParams 获取函数参数
func (self Type) GetFuncParams() []Type {
	if !self.IsFunc() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetFuncParams()
	}
	return self.elems[1:]
}

// GetFuncVarArg 获取函数是否是不定参数
func (self Type) GetFuncVarArg() bool {
	if !self.IsFunc() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetFuncVarArg()
	}
	return self.varArg
}

// GetArraySize 获取数组大小
func (self Type) GetArraySize() uint {
	if !self.IsArray() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetArraySize()
	}
	return self.width
}

// GetArrayElem 获取数组元素
func (self Type) GetArrayElem() Type {
	if !self.IsArray() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetArrayElem()
	}
	return self.elems[0]
}

// GetStructElems 获取结构体元素
func (self Type) GetStructElems() []Type {
	if !self.IsStruct() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetStructElems()
	}
	return self.elems
}

// GetUnionElems 获取联合元素
func (self Type) GetUnionElems() []Type {
	if !self.IsUnion() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetUnionElems()
	}
	return self.elems
}

// GetUnionMaxElemAlign 获取联合最大字段对齐
func (self Type) GetUnionMaxElemAlign() uint {
	if !self.IsUnion() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetUnionMaxElemAlign()
	}
	var maxAlign uint = 0
	for _, e := range self.elems {
		a := e.Align()
		if a > maxAlign {
			maxAlign = a
		}
	}
	return maxAlign
}

// GetUnionMaxElemSize 获取联合最大字段大小
func (self Type) GetUnionMaxElemSize() uint {
	if !self.IsUnion() {
		panic("unreachable")
	}
	if self.IsAlias() {
		return self.GetAliasTarget().GetUnionMaxElemSize()
	}
	var maxSize uint = 0
	for _, e := range self.elems {
		es := e.Size()
		if es > maxSize {
			maxSize = es
		}
	}
	return maxSize
}

// GetAlias 获取别名
func (self Type) GetAlias() *Alias {
	if !self.IsAlias() {
		panic("unreachable")
	}
	return self.alias
}

// GetAliasTarget 获取别名目标类型
func (self Type) GetAliasTarget() Type {
	if !self.IsAlias() {
		panic("unreachable")
	}
	return self.alias.Target
}

// GetDeepAliasTarget 获取深度别名目标类型
func (self Type) GetDeepAliasTarget() Type {
	if !self.IsAlias() {
		panic("unreachable")
	}

	cursor := self
	for cursor.IsAlias() {
		cursor = cursor.GetAliasTarget()
	}
	return cursor
}

// Equal 比较
func (self Type) Equal(dst Type) bool {
	for self.IsAlias() && self.GetAliasTarget().IsAlias() {
		self = self.GetAliasTarget()
	}
	for dst.IsAlias() && dst.GetAliasTarget().IsAlias() {
		dst = dst.GetAliasTarget()
	}
	if self.IsAlias() && dst.IsAlias() && self.alias.Name == dst.alias.Name {
		return true
	}

	if self.IsAlias() {
		self = self.GetAliasTarget()
	}
	if dst.IsAlias() {
		dst = dst.GetAliasTarget()
	}

	if self.Kind != dst.Kind {
		return false
	}

	switch self.Kind {
	case TVoid, TBool:
		return true
	case TSint, TUint, TFloat:
		return self.width == dst.width
	case TArray:
		return self.width == dst.width && self.elems[0].Equal(dst.elems[0])
	case TFunc:
		if self.varArg != dst.varArg {
			return false
		}
		fallthrough
	case TPtr, TStruct, TUnion:
		if len(self.elems) != len(dst.elems) {
			return false
		}
		for i, e := range self.elems {
			if !e.Equal(dst.elems[i]) {
				return false
			}
		}
		return true
	default:
		panic("unreachable")
	}
}

// Align 获取对齐（byte）
func (self Type) Align() uint {
	switch self.Kind {
	case TVoid:
		panic("unreachable")
	case TBool:
		return 1
	case TSint, TUint:
		if self.width == 0 {
			return utils.PtrByte
		}
		return self.width
	case TFloat:
		return self.width
	case TPtr, TFunc:
		return utils.PtrByte
	case TArray:
		return self.elems[0].Align()
	case TStruct:
		var align uint
		for _, e := range self.elems {
			align = uint(math.Max(float64(e.Align()), float64(align)))
		}
		return align
	case TUnion:
		return self.GetUnionMaxElemAlign()
	case TAlias:
		return self.alias.Target.Align()
	default:
		panic("unreachable")
	}
}

// Size 获取大小（byte）
func (self Type) Size() uint {
	switch self.Kind {
	case TVoid:
		panic("unreachable")
	case TBool:
		return 1
	case TSint, TUint:
		if self.width == 0 {
			return utils.PtrByte
		}
		return self.width
	case TFloat:
		return self.width
	case TPtr, TFunc:
		return utils.PtrByte
	case TArray:
		return self.width * self.elems[0].Size()
	case TStruct:
		var offset uint
		for _, e := range self.elems {
			es := e.Size()
			offset += es
			offset = utils.AlignTo(offset, es)
		}
		return utils.AlignTo(offset, self.Align())
	case TUnion:
		return utils.AlignTo(self.GetUnionMaxElemSize(), self.GetUnionMaxElemAlign())
	case TAlias:
		return self.alias.Target.Size()
	default:
		panic("unreachable")
	}
}
