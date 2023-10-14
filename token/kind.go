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

	FUNC
)

var kindNames = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	IDENT:   "ident",
	LPA:     "lpa",
	RPA:     "rpa",
	LBR:     "lbr",
	RBR:     "rbr",
	FUNC:    "func",
}

// Lookup 区分标识符和关键字
func Lookup(s string) Kind {
	switch s {
	case "func":
		return FUNC
	default:
		return IDENT
	}
}

func (self Kind) String() string {
	return kindNames[self]
}
