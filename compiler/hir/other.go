package hir

import (
	"github.com/kkkunny/stl/container/linkedlist"
	stlslices "github.com/kkkunny/stl/slices"
)

// 语义分析结果
type Result struct {
	BuildinTypes struct {
		Isize, I8, I16, I32, I64 Type
		Usize, U8, U16, U32, U64 Type
		F32, F64                 Type
		Bool, Str                Type
		Default                  *Trait
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
