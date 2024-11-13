package types

// AliasType 类型别名
type AliasType interface {
	Type
	Name() string
	Target() Type
	Alias()
}
