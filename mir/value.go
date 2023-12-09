package mir

import "fmt"

type Value interface {
	Name()string
	Type()Type
}

// Param 参数
type Param struct {
	index uint
	t Type
}

func newParam(i uint, t Type)*Param{
	return &Param{
		index: i,
		t: t,
	}
}

func (self *Param) Type()Type{
	return self.t
}

func (self *Param) Name()string{
	return fmt.Sprintf("p%d", self.index)
}

func (self *Param) Define()string{
	return fmt.Sprintf("%s %s", self.Type(), self.Name())
}
