package token

// Kind token类型
type Kind uint8

const (
	ILLEGAL Kind = iota
	EOF

	IDENT
	INTEGER
	FLOAT

	ASS

	AND
	OR
	XOR
	NOT
	SHL
	SHR

	ADD
	SUB
	MUL
	DIV
	REM

	EQ
	NE
	LT
	GT
	LE
	GE

	LPA
	RPA
	LBA
	RBA
	LBR
	RBR

	SEM
	COM
	DOT
	COL

	_KeywordBegin
	FUNC
	RETURN
	TRUE
	FALSE
	AS
	STRUCT
	LET
	IF
	ELSE
	_KeywordEnd
)

var kindNames = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	IDENT:   "ident",
	INTEGER: "integer",
	FLOAT:   "float",
	ASS:     "ass",
	AND:     "and",
	OR:      "or",
	XOR:     "xor",
	NOT:     "not",
	SHL:     "shl",
	SHR:     "shr",
	ADD:     "add",
	SUB:     "sub",
	MUL:     "mul",
	DIV:     "div",
	REM:     "rem",
	EQ:      "eq",
	NE:      "ne",
	LT:      "lt",
	GT:      "gt",
	LE:      "le",
	GE:      "ge",
	LPA:     "lpa",
	RPA:     "rpa",
	LBA:     "lba",
	RBA:     "rba",
	LBR:     "lbr",
	RBR:     "rbr",
	SEM:     "sem",
	COM:     "com",
	DOT:     "dot",
	COL:     "col",
	FUNC:    "func",
	RETURN:  "return",
	TRUE:    "true",
	FALSE:   "false",
	AS:      "as",
	STRUCT:  "struct",
	LET:     "let",
	ELSE:    "else",
	IF:      "if",
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
		return 8
	case ADD, SUB:
		return 7
	case SHL, SHR:
		return 6
	case LT, GT, LE, GE:
		return 5
	case EQ, NE:
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
