package token

// Kind token类型
type Kind uint8

const (
	ILLEGAL Kind = iota
	EOF

	IDENT
	INTEGER
	FLOAT

	AND
	OR
	XOR

	ADD
	SUB
	MUL
	DIV
	REM

	LT
	GT
	LE
	GE

	LPA
	RPA
	LBR
	RBR

	SEM

	_KeywordBegin
	FUNC
	RETURN
	TRUE
	FALSE
	AS
	_KeywordEnd
)

var kindNames = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	IDENT:   "ident",
	INTEGER: "integer",
	FLOAT:   "float",
	AND:     "and",
	OR:      "or",
	XOR:     "xor",
	ADD:     "add",
	SUB:     "sub",
	MUL:     "mul",
	DIV:     "div",
	REM:     "rem",
	LPA:     "lpa",
	LT:      "lt",
	GT:      "gt",
	LE:      "le",
	GE:      "ge",
	RPA:     "rpa",
	LBR:     "lbr",
	RBR:     "rbr",
	SEM:     "sem",
	FUNC:    "func",
	RETURN:  "return",
	TRUE:    "true",
	FALSE:   "false",
	AS:      "as",
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

// Priority 运算符优先级
func (self Kind) Priority() uint8 {
	switch self {
	case MUL, DIV, REM:
		return 6
	case ADD, SUB:
		return 5
	case LT, GT, LE, GE:
		return 4
	case AND:
		return 3
	case XOR:
		return 2
	case OR:
		return 1
	default:
		return 0
	}
}
