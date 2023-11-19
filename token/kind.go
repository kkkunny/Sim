package token

// Kind token类型
type Kind uint8

const (
	ILLEGAL Kind = iota
	EOF

	IDENT
	INTEGER
	FLOAT
	CHAR
	STRING

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

	LAND
	LOR

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
	SCOPE

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
	MUT
	LOOP
	BREAK
	CONTINUE
	FOR
	IN
	IMPORT
	_KeywordEnd
)

var kindNames = [...]string{
	ILLEGAL:  "illegal",
	EOF:      "eof",
	IDENT:    "ident",
	INTEGER:  "integer",
	FLOAT:    "float",
	CHAR:     "char",
	STRING:   "string",
	ASS:      "ass",
	AND:      "and",
	OR:       "or",
	XOR:      "xor",
	NOT:      "not",
	SHL:      "shl",
	SHR:      "shr",
	ADD:      "add",
	SUB:      "sub",
	MUL:      "mul",
	DIV:      "div",
	REM:      "rem",
	EQ:       "eq",
	NE:       "ne",
	LT:       "lt",
	GT:       "gt",
	LE:       "le",
	GE:       "ge",
	LAND:     "logic and",
	LOR:      "logic or",
	LPA:      "lpa",
	RPA:      "rpa",
	LBA:      "lba",
	RBA:      "rba",
	LBR:      "lbr",
	RBR:      "rbr",
	SEM:      "sem",
	COM:      "com",
	DOT:      "dot",
	COL:      "col",
	SCOPE:    "scope",
	FUNC:     "func",
	RETURN:   "return",
	TRUE:     "true",
	FALSE:    "false",
	AS:       "as",
	STRUCT:   "struct",
	LET:      "let",
	ELSE:     "else",
	IF:       "if",
	MUT:      "mut",
	BREAK:    "break",
	CONTINUE: "continue",
	LOOP:     "loop",
	FOR:      "for",
	IN:       "in",
	IMPORT:   "import",
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
		return 11
	case ADD, SUB:
		return 10
	case SHL, SHR:
		return 9
	case LT, GT, LE, GE:
		return 8
	case EQ, NE:
		return 7
	case AND:
		return 6
	case XOR:
		return 5
	case OR:
		return 4
	case LAND:
		return 3
	case LOR:
		return 2
	case ASS:
		return 1
	default:
		return 0
	}
}
