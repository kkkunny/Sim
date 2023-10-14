package token

// Kind token类型
type Kind uint8

const (
	Illegal Kind = iota
	Eof
)

var kindNames = [...]string{
	Illegal: "illegal",
	Eof:     "eof",
}

func (self Kind) String() string {
	return kindNames[self]
}
