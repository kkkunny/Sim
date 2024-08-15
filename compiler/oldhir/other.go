package oldhir

import (
	"github.com/kkkunny/stl/container/linkedlist"
	stlslices "github.com/kkkunny/stl/container/slices"
)

// Result 语义分析结果
type Result struct {
	BuildinTypes struct {
		Isize, I8, I16, I32, I64 Type
		Usize, U8, U16, U32, U64 Type
		F32, F64                 Type
		Bool, Str                Type
		Default                  *Trait
		Copy                     *Trait
		Add, Sub, Mul, Div, Rem  *Trait
		And, Or, Xor, Shl, Shr   *Trait
		Eq, Lt, Gt, Land, Lor    *Trait
		Neg, Not                 *Trait
	}
	Globals linkedlist.LinkedList[Global]
}

// FuncDecl 函数声明
type FuncDecl struct {
	Name   string
	Params []*Param
	Ret    Type
}

func (self FuncDecl) GetType() *FuncType {
	return NewFuncType(self.Ret, stlslices.Map(self.Params, func(_ int, e *Param) Type {
		return e.GetType()
	})...)
}

// CallableDef 可调用定义
type CallableDef interface {
	GetFuncType() *FuncType
}
