package values

// Ident 标识符
type Ident interface {
	Value
	GetName() (string, bool)
}
