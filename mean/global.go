package mean

// Global 全局
type Global interface {
	global()
}

// FuncDef 函数定义
type FuncDef struct {
	Name string
	Ret  Type
	Body *Block
}

func (self *FuncDef) global() {}

func (self *FuncDef) stmt() {}

func (self *FuncDef) GetType() Type {
	return &FuncType{Ret: self.Ret}
}

func (self *FuncDef) ident() {}
