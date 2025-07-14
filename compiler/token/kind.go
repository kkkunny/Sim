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
	COMMENT

	ASS

	AND
	AND_WITH_MUT
	OR
	XOR
	NOT
	SHL
	SHR

	ANDASS
	ORASS
	XORASS
	SHLASS
	SHRASS

	ADD
	SUB
	MUL
	DIV
	REM

	ADDASS
	SUBASS
	MULASS
	DIVASS
	REMASS

	EQ
	NE
	LT
	GT
	LE
	GE

	LAND
	LOR

	LANDASS
	LORASS

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
	AT
	QUE
	ARROW

	_KeywordBegin
	FUNC
	RETURN
	AS
	STRUCT
	LET
	IF
	ELSE
	MUT
	WHILE
	BREAK
	CONTINUE
	FOR
	IN
	IMPORT
	IS
	PUBLIC
	TYPE
	OTHER
	MATCH
	TRAIT
	ENUM
	CASE
	_KeywordEnd
)

var kindNames = [...]string{
	ILLEGAL:      "illegal",
	EOF:          "eof",
	IDENT:        "ident",
	INTEGER:      "integer",
	FLOAT:        "float",
	CHAR:         "char",
	STRING:       "string",
	COMMENT:      "comment",
	ASS:          "ass",
	AND:          "and",
	AND_WITH_MUT: "and",
	OR:           "or",
	XOR:          "xor",
	NOT:          "not",
	SHL:          "shl",
	SHR:          "shr",
	ANDASS:       "and ass",
	ORASS:        "or ass",
	XORASS:       "xor ass",
	SHLASS:       "shl ass",
	SHRASS:       "shr ass",
	ADD:          "add",
	SUB:          "sub",
	MUL:          "mul",
	DIV:          "div",
	REM:          "rem",
	ADDASS:       "and ass",
	SUBASS:       "sub ass",
	MULASS:       "mul ass",
	DIVASS:       "div ass",
	REMASS:       "rem ass",
	EQ:           "eq",
	NE:           "ne",
	LT:           "lt",
	GT:           "gt",
	LE:           "le",
	GE:           "ge",
	LAND:         "logic and",
	LOR:          "logic or",
	LANDASS:      "land ass",
	LORASS:       "lor ass",
	LPA:          "lpa",
	RPA:          "rpa",
	LBA:          "lba",
	RBA:          "rba",
	LBR:          "lbr",
	RBR:          "rbr",
	SEM:          "sem",
	COM:          "com",
	DOT:          "dot",
	COL:          "col",
	SCOPE:        "scope",
	AT:           "at",
	QUE:          "que",
	ARROW:        "arrow",
	FUNC:         "func",
	RETURN:       "return",
	AS:           "as",
	STRUCT:       "struct",
	LET:          "let",
	ELSE:         "else",
	IF:           "if",
	MUT:          "mut",
	BREAK:        "break",
	CONTINUE:     "continue",
	WHILE:        "while",
	FOR:          "for",
	IN:           "in",
	IMPORT:       "import",
	IS:           "is",
	PUBLIC:       "pub",
	TYPE:         "type",
	OTHER:        "other",
	MATCH:        "match",
	TRAIT:        "trait",
	ENUM:         "enum",
	CASE:         "case",
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
	case AND, AND_WITH_MUT:
		return 6
	case XOR:
		return 5
	case OR:
		return 4
	case LAND:
		return 3
	case LOR:
		return 2
	case ASS, ANDASS, ORASS, XORASS, SHLASS, SHRASS, ADDASS, SUBASS, MULASS, DIVASS, REMASS, LANDASS, LORASS:
		return 1
	default:
		return 0
	}
}
