package mean

var (
	Empty = &EmptyType{}

	Isize = &IntType{Bits: 64}
)

// Type 类型
type Type interface {
	Equal(dst Type) bool
}

// EmptyType 空类型
type EmptyType struct{}

func (self *EmptyType) Equal(dst Type) bool {
	_, ok := dst.(*EmptyType)
	return ok
}

type IntType struct {
	Bits uint
}

func (self *IntType) Equal(dst Type) bool {
	t, ok := dst.(*IntType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}
