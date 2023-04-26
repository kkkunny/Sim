package mir

// Value 值
type Value interface {
	GetType() Type
}

// Param 参数
type Param struct {
	Type Type
}

func (self *Function) newParam(t Type) *Param {
	return &Param{
		Type: t,
	}
}
func (self Param) GetType() Type {
	return NewTypePtr(self.Type)
}
