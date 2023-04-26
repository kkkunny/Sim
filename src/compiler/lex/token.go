package lex

import (
	"fmt"

	"github.com/kkkunny/Sim/src/compiler/utils"
)

// TokenKind token kind
type TokenKind uint

const (
	ILLEGAL TokenKind = iota // 非法
	EOF                      // 结束符
	COMMENT                  // 注释

	IDENT  // 标识符
	Attr   // 属性
	INT    // 整数
	FLOAT  // 浮点数
	CHAR   // 字符
	STRING // 字符串
	NULL   // 空指针

	ASS // =
	ADS // +=
	SUS // -=
	MUS // *=
	DIS // /=
	MOS // %=
	ANS // &=
	ORS // |=
	XOS // ^=
	SLS // <<=
	SRS // >>=

	ADD // +
	SUB // -
	MUL // *
	DIV // /
	MOD // %

	AND // &
	OR  // |
	XOR // ^
	SHL // <<
	SHR // >>

	EQ // ==
	NE // !=
	LT // <
	LE // <=
	GT // >
	GE // >=

	LAN // &&
	LOR // ||

	LENOF   // len
	TYPEOF  // typename
	SIZEOF  // sizeof
	ALIGNOF // alignof
	PTROF   // ptrof
	VALOF   // valof

	LPA // 左小括号
	RPA // 右小括号
	LBA // 左中括号
	RBA // 右中括号
	LBR // 左大括号
	RBR // 右大括号

	SEM // 分隔符
	COL // :
	CLL // ::
	NOT // !
	NEG // ~
	COM // ,
	DOT // .
	ELL // ..
	QUO // ?
	CLT // ::<

	FUNC     // func
	RETURN   // return
	TRUE     // true
	FALSE    // false
	STRUCT   // struct
	IF       // if
	ELSE     // else
	FOR      // for
	BREAK    // break
	CONTINUE // continue
	AS       // as
	TYPE     // type
	IMPORT   // import
	PUB      // pub
	LET      // let
	MATCH    // match
	CASE     // case
	DEFAULT  // default
	MUT      // mut
	UNSAFE   // unsafe
)

var tokenKindStr = [...]string{
	ILLEGAL: "illegal",
	EOF:     "eof",
	COMMENT: "comment",

	IDENT:  "ident",
	Attr:   "attr",
	INT:    "int",
	FLOAT:  "float",
	CHAR:   "char",
	STRING: "string",
	NULL:   "null",

	ASS: "=",
	ADS: "+=",
	SUS: "-=",
	MUS: "*=",
	DIS: "/=",
	MOS: "%=",
	ANS: "&=",
	ORS: "|=",
	XOS: "^=",
	SLS: "<<=",
	SRS: ">>=",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	MOD: "%",

	AND: "&",
	OR:  "|",
	XOR: "^",
	SHL: "<<",
	SHR: ">>",

	EQ: "==",
	NE: "!=",
	LT: "<",
	LE: "<=",
	GT: ">",
	GE: ">=",

	LAN: "&&",
	LOR: "||",

	LENOF:   "lenof",
	TYPEOF:  "typeof",
	SIZEOF:  "sizeof",
	ALIGNOF: "alignof",
	PTROF:   "ptrof",
	VALOF:   "valof",

	LPA: "(",
	RPA: ")",
	LBA: "[",
	RBA: "]",
	LBR: "{",
	RBR: "}",

	SEM: ";",
	COL: ":",
	CLL: "::",
	NOT: "!",
	NEG: "~",
	COM: ",",
	DOT: ".",
	ELL: "..",
	QUO: "?",
	CLT: "::<",

	FUNC:     "func",
	RETURN:   "return",
	TRUE:     "true",
	FALSE:    "false",
	STRUCT:   "struct",
	IF:       "if",
	ELSE:     "else",
	FOR:      "for",
	BREAK:    "break",
	CONTINUE: "continue",
	AS:       "as",
	TYPE:     "type",
	IMPORT:   "import",
	PUB:      "pub",
	LET:      "let",
	MATCH:    "match",
	CASE:     "case",
	DEFAULT:  "default",
	MUT:      "mut",
	UNSAFE:   "unsafe",
}

// LookUp 区分标识符和关键字
func LookUp(s string) TokenKind {
	switch s {
	case "lenof":
		return LENOF
	case "typeof":
		return TYPEOF
	case "sizeof":
		return SIZEOF
	case "alignof":
		return ALIGNOF
	case "ptrof":
		return PTROF
	case "valof":
		return VALOF
	case "func":
		return FUNC
	case "return":
		return RETURN
	case "true":
		return TRUE
	case "false":
		return FALSE
	case "struct":
		return STRUCT
	case "if":
		return IF
	case "else":
		return ELSE
	case "for":
		return FOR
	case "break":
		return BREAK
	case "continue":
		return CONTINUE
	case "as":
		return AS
	case "type":
		return TYPE
	case "null":
		return NULL
	case "import":
		return IMPORT
	case "pub":
		return PUB
	case "let":
		return LET
	case "match":
		return MATCH
	case "case":
		return CASE
	case "default":
		return DEFAULT
	case "mut":
		return MUT
	case "unsafe":
		return UNSAFE
	default:
		return IDENT
	}
}

func (self TokenKind) String() string {
	return tokenKindStr[self]
}

// Priority 获取运算符优先级
func (self TokenKind) Priority() uint8 {
	switch self {
	case MUL, DIV, MOD:
		return 6
	case ADD, SUB:
		return 5
	case EQ, NE, LT, LE, GT, GE:
		return 4
	case AND, OR, XOR, SHL, SHR:
		return 3
	case LAN, LOR:
		return 2
	case ASS, ADS, SUS, MUS, DIS, MOS, ANS, ORS, XOS, SLS, SRS:
		return 1
	default:
		return 0
	}
}

// Token token
type Token struct {
	Pos    utils.Position // 位置
	Kind   TokenKind      // kind
	Source string         // 源码
}

func (self Token) String() string {
	return fmt.Sprintf("<%s: %s>", self.Kind, self.Source)
}
