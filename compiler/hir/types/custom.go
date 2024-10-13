package types

// CustomType 自定义类型，一定继承自global.CustomTypeWrap
type CustomType interface {
	Type
	Name() string
	Target() Type
	Custom()
}
