package hir

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

// GenericInst 泛型实例化
type GenericInst interface {
	GetGenericParams()[]*GenericIdentType
	GetGenericArgs()[]Type
}
