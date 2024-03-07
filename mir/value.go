package mir

import "fmt"

type Value interface {
	Name() string
	Type() Type
	ReadRefValues() []Value
}

// Param 参数
type Param struct {
	index uint
	t     Type
}

func newParam(i uint, t Type) *Param {
	return &Param{
		index: i,
		t:     t,
	}
}

func (self *Param) Type() Type {
	return self.t.Context().NewPtrType(self.t)
}

func (self *Param) ValueType() Type {
	return self.t
}

func (self *Param) Name() string {
	return fmt.Sprintf("p%d", self.index)
}

func (self *Param) Define() string {
	return fmt.Sprintf("%s %s", self.ValueType(), self.Name())
}

func (self *Param) ReadRefValues() []Value {
	return nil
}
