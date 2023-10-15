package mean

var (
	Empty = &EmptyType{}
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
