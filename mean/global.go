package mean

import "github.com/samber/lo"

// Global 全局
type Global interface {
	global()
}

// FuncDef 函数定义
type FuncDef struct {
	Name   string
	Params []*Param
	Ret    Type
	Body   *Block
}

func (self *FuncDef) global() {}

func (self *FuncDef) stmt() {}

func (self *FuncDef) GetType() Type {
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: params,
	}
}

func (self *FuncDef) ident() {}
