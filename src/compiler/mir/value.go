package mir

import "fmt"

// Value 值
type Value interface {
	GetName() string
	GetType() Type
}

// Param 参数
type Param struct {
	index uint
	Type  Type
}

func (self *Function) newParam(i uint, t Type) *Param {
	return &Param{
		index: i,
		Type:  t,
	}
}
func (self Param) GetName() string {
	return fmt.Sprintf("p%d", self.index)
}
func (self Param) GetType() Type {
	return NewTypePtr(self.Type)
}
