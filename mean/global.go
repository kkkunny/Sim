package mean

// Global 全局
type Global interface {
	global()
}

// FuncDef 函数定义
type FuncDef struct {
	Name string
	Body *Block
}

func (self *FuncDef) global() {}
