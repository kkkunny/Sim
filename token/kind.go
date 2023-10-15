package token

// Kind token类型
type Kind uint8

const (
	ILLEGAL Kind = iota
	EOF

	IDENT
	INTEGER
	FLOAT

	LPA
	RPA
	LBR
	RBR

	SEM

	_KeywordBegin
	FUNC
	RETURN
	_KeywordEnd
)

var kindNames = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	IDENT:   "ident",
	INTEGER: "integer",
	FLOAT:   "float",
	LPA:     "lpa",
	RPA:     "rpa",
	LBR:     "lbr",
	RBR:     "rbr",
	SEM:     "sem",
	FUNC:    "func",
	RETURN:  "return",
}

// Lookup 区分标识符和关键字
func Lookup(s string) Kind {
	kindStrMap := make(map[string]Kind, _KeywordEnd-_KeywordBegin-1)
	for k, s := range kindNames[_KeywordBegin+1 : _KeywordEnd] {
		kindStrMap[s] = _KeywordBegin + 1 + Kind(k)
	}
	keyword, ok := kindStrMap[s]
	if !ok {
		return IDENT
	}
	return keyword
}

func (self Kind) String() string {
	return kindNames[self]
}
