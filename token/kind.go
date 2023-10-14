package token

// Kind token类型
type Kind uint8

const (
	ILLEGAL Kind = iota
	EOF

	IDENT

	LPA
	RPA
	LBR
	RBR
)

var kindNames = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	IDENT:   "ident",
	LPA:     "lpa",
	RPA:     "rpa",
	LBR:     "lbr",
	RBR:     "rbr",
}

func (self Kind) String() string {
	return kindNames[self]
}
