package token

// Kind token类型
type Kind uint8

const (
	Illegal Kind = iota
)

var kindNames = [...]string{
	Illegal: "illegal",
}

func (self Kind) String() string {
	return kindNames[self]
}
