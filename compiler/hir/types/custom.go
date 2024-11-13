package types

// CustomType 自定义类型
type CustomType interface {
	Type
	Name() string
	Target() Type
	Custom()
}
