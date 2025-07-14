package types

// AliasType 类型别名
type AliasType interface {
	TypeDef
	Alias()
}
