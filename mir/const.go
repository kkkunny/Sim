package mir

type Const interface {
	Value
	constant()
}
