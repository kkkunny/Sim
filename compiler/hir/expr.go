package hir

// Expr 表达式
type Expr interface {
	GetType() Type
	Mutable() bool
}

// Ident 标识符
type Ident interface {
	Expr
	GetName()string
}

// Variable 变量
type Variable interface {
	Ident
	variable()
}

// VarDecl 变量声明
type VarDecl struct {
	Mut  bool
	Type Type
	Name string
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

func (*VarDecl) variable(){}

// Param 参数
type Param struct {
	VarDecl
}
