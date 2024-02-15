package hir

import (
	"github.com/kkkunny/stl/container/linkedlist"
)

// 语义分析结果
type Result struct {
	BuildinTypes struct {
		Isize, I8, I16, I32, I64 Type
		Usize, U8, U16, U32, U64 Type
		F32, F64                 Type
		Bool, Str                Type
	}
	Globals linkedlist.LinkedList[Global]
}

// FuncDecl 函数声明
type FuncDecl struct {
	Name   string
	Params []*Param
	Ret    Type
}
