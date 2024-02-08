package hir

import "github.com/kkkunny/stl/container/linkedlist"

// 语义分析结果
type Result struct{
	BuildinTypes struct{
		Isize, I8, I16, I32, I64 Type
		Usize, U8, U16, U32, U64 Type
		F32, F64 Type
		Bool, Str Type
	}
	Globals linkedlist.LinkedList[Global]
}

// Variable 变量
type Variable interface {
	Ident
	GetName()string
}

// VarDecl 变量声明
type VarDecl struct {
	Mut        bool
	Type       Type
	Name       string
}

func (self *VarDecl) GetName() string {
	return self.Name
}

func (self *VarDecl) GetType() Type {
	return self.Type
}

func (self *VarDecl) Mutable() bool {
	return self.Mut
}

// Param 参数
type Param struct {
	VarDecl
}

func (*Param) stmt() {}

func (*Param) ident() {}
