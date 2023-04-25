package lex

import (
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
	stlutil "github.com/kkkunny/stl/util"
)

// Lexer 词法分析器
type Lexer struct {
	file   stlos.Path // 源文件路径
	reader Reader     // 读取器
	pos    uint       // 当前下标(byte)
	ch     rune       // 当前字符
	row    uint       // 当前行数
	col    uint       // 当前列数
}

// NewLexer 新建词法分析器
func NewLexer(fp stlos.Path, reader Reader) *Lexer {
	lexer := &Lexer{
		file:   fp,
		reader: reader,
		row:    1,
	}
	lexer.next()
	return lexer
}

// GetFilepath 获取源文件路径
func (self *Lexer) GetFilepath() stlos.Path {
	return self.file
}

// 下一个
func (self *Lexer) next() rune {
	ch, size, err := self.reader.ReadRune()
	if err != nil && errors.Is(err, io.EOF) {
		self.ch = 0
		return 0
	} else if err != nil {
		panic(err)
	} else {
		self.ch = ch
		self.pos += uint(size)
		if ch == '\n' {
			self.row++
			self.col = 0
		} else {
			self.col++
		}
		return ch
	}
}

// 获取下一个
func (self *Lexer) peek() rune {
	ch, _, err := self.reader.ReadRune()
	if err != nil && errors.Is(err, io.EOF) {
		return 0
	} else if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
	stlutil.MustValue(self.reader.Seek(-1, io.SeekCurrent))
	return ch
}

// 跳过空白
func (self *Lexer) skipWhiteSpace() {
	for self.ch == ' ' || self.ch == '\t' {
		self.next()
	}
}

// 扫描标识符
func (self *Lexer) scanIdent() Token {
	pos := utils.NewPosition(self.file)
	pos.SetBegin(self.pos, self.row, self.col)

	var buf strings.Builder
	for self.ch == '_' || unicode.IsLetter(self.ch) || utils.IsNumber(self.ch) {
		buf.WriteRune(self.ch)
		pos.SetEnd(self.pos, self.row, self.col)
		self.next()
	}
	s := buf.String()

	return Token{
		Pos:    pos,
		Kind:   LookUp(s),
		Source: s,
	}
}

// 扫描属性
func (self *Lexer) scanAttr() Token {
	pos := utils.NewPosition(self.file)
	pos.SetBegin(self.pos, self.row, self.col)
	pos.SetEnd(self.pos, self.row, self.col)
	self.next()

	if self.ch != '_' && !unicode.IsLetter(self.ch) {
		return Token{
			Pos:    pos,
			Kind:   ILLEGAL,
			Source: "@",
		}
	}

	ident := self.scanIdent()
	pos = utils.MixPosition(pos, ident.Pos)
	s := "@" + ident.Source

	return Token{
		Pos:    pos,
		Kind:   stlutil.Ternary(ident.Kind == IDENT, Attr, ILLEGAL),
		Source: s,
	}
}

// 扫描数字
func (self *Lexer) scanNumber() Token {
	pos := utils.NewPosition(self.file)
	pos.SetBegin(self.pos, self.row, self.col)

	var buf strings.Builder
	var point int
	for self.ch == '.' || utils.IsNumber(self.ch) {
		if self.ch == '.' {
			point++
		}
		buf.WriteRune(self.ch)
		pos.SetEnd(self.pos, self.row, self.col)
		self.next()
	}
	s := buf.String()

	if point > 1 || s[len(s)-1] == '.' {
		return Token{
			Pos:    pos,
			Kind:   ILLEGAL,
			Source: s,
		}
	}

	return Token{
		Pos:    pos,
		Kind:   stlutil.Ternary(point == 0, INT, FLOAT),
		Source: s,
	}
}

// 扫描字符
func (self *Lexer) scanChar() Token {
	pos := utils.NewPosition(self.file)
	pos.SetBegin(self.pos, self.row, self.col)

	var buf strings.Builder
	var lastRune rune
	for self.ch != 0 {
		ch := self.ch
		buf.WriteRune(ch)
		pos.SetEnd(self.pos, self.row, self.col)
		self.next()
		if buf.Len() != 1 && ch == '\'' && lastRune != '\\' {
			break
		}
		lastRune = ch
	}
	s := buf.String()
	ss := utils.ParseEscapeCharacter(s, `\'`, `'`)
	runes := []rune(ss)

	if len(runes) != 3 || runes[0] != '\'' || runes[2] != '\'' {
		return Token{
			Pos:    pos,
			Kind:   ILLEGAL,
			Source: s,
		}
	}

	return Token{
		Pos:    pos,
		Kind:   CHAR,
		Source: ss,
	}
}

// 扫描字符串
func (self *Lexer) scanString() Token {
	pos := utils.NewPosition(self.file)
	pos.SetBegin(self.pos, self.row, self.col)

	var buf strings.Builder
	var lastRune rune
	for self.ch != 0 {
		ch := self.ch
		buf.WriteRune(ch)
		pos.SetEnd(self.pos, self.row, self.col)
		self.next()
		if buf.Len() != 1 && ch == '"' && lastRune != '\\' {
			break
		}
		lastRune = ch
	}
	s := buf.String()

	if s[0] != '"' || s[len(s)-1] != '"' {
		return Token{
			Pos:    pos,
			Kind:   ILLEGAL,
			Source: s,
		}
	}

	s = utils.ParseEscapeCharacter(s, `\"`, `"`)
	return Token{
		Pos:    pos,
		Kind:   STRING,
		Source: s,
	}
}

// Scan 词法分析
func (self *Lexer) Scan() Token {
	self.skipWhiteSpace()

	if self.ch == '_' || unicode.IsLetter(self.ch) {
		return self.scanIdent()
	} else if self.ch == '@' {
		return self.scanAttr()
	} else if utils.IsNumber(self.ch) {
		return self.scanNumber()
	} else if self.ch == '\'' {
		return self.scanChar()
	} else if self.ch == '"' {
		return self.scanString()
	} else {
		pos := utils.NewPosition(self.file)
		pos.SetBegin(self.pos, self.row, self.col)
		kind := ILLEGAL

		var buf strings.Builder
		switch self.ch {
		case 0:
			kind = EOF
			buf.WriteRune(self.ch)
		case '=':
			kind = ASS
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = EQ
			}
		case '+':
			kind = ADD
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = ADS
			}
		case '-':
			kind = SUB
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = SUS
			}
		case '*':
			kind = MUL
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = MUS
			}
		case '/':
			kind = DIV
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = DIS
			} else if self.peek() == '/' {
				buf.WriteRune(self.next())
				for nc := self.peek(); nc != '\n' && nc != 0; nc = self.peek() {
					buf.WriteRune(self.next())
				}
				kind = COMMENT
			} else if self.peek() == '*' {
				buf.WriteRune(self.next())
				self.next()
				var lastRune rune
				for {
					if self.ch == 0 {
						pos.SetEnd(self.pos, self.row, self.col)
						return Token{
							Pos:    pos,
							Kind:   ILLEGAL,
							Source: buf.String(),
						}
					}
					buf.WriteRune(self.ch)
					if lastRune == '*' && self.ch == '/' {
						break
					}
					lastRune = self.ch
					self.next()
				}
				kind = COMMENT
			}
		case '%':
			kind = MOD
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = MOS
			}
		case '&':
			kind = AND
			buf.WriteRune(self.ch)
			if self.peek() == '&' {
				buf.WriteRune(self.next())
				kind = LAN
			} else if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = ANS
			}
		case '|':
			kind = OR
			buf.WriteRune(self.ch)
			if self.peek() == '|' {
				buf.WriteRune(self.next())
				kind = LOR
			} else if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = ORS
			}
		case '^':
			kind = XOR
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = XOS
			}
		case '<':
			kind = LT
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = LE
			} else if self.peek() == '<' {
				buf.WriteRune(self.next())
				kind = SHL
				if self.peek() == '=' {
					buf.WriteRune(self.next())
					kind = SLS
				}
			}
		case '>':
			kind = GT
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = GE
			} else if self.peek() == '>' {
				buf.WriteRune(self.next())
				kind = SHR
				if self.peek() == '=' {
					buf.WriteRune(self.next())
					kind = SRS
				}
			}
		case '(':
			kind = LPA
			buf.WriteRune(self.ch)
		case ')':
			kind = RPA
			buf.WriteRune(self.ch)
		case '[':
			kind = LBA
			buf.WriteRune(self.ch)
		case ']':
			kind = RBA
			buf.WriteRune(self.ch)
		case '{':
			kind = LBR
			buf.WriteRune(self.ch)
		case '}':
			kind = RBR
			buf.WriteRune(self.ch)
		case ';', '\n':
			kind = SEM
			buf.WriteRune(';')
		case ':':
			kind = COL
			buf.WriteRune(self.ch)
			if self.peek() == ':' {
				buf.WriteRune(self.next())
				kind = CLL
				if self.peek() == '<' {
					buf.WriteRune(self.next())
					kind = CLT
				}
			}
		case '!':
			kind = NOT
			buf.WriteRune(self.ch)
			if self.peek() == '=' {
				buf.WriteRune(self.next())
				kind = NE
			}
		case '~':
			kind = NEG
			buf.WriteRune(self.ch)
		case '?':
			kind = QUO
			buf.WriteRune(self.ch)
		case ',':
			kind = COM
			buf.WriteRune(self.ch)
		case '.':
			kind = DOT
			buf.WriteRune(self.ch)
			if self.peek() == '.' {
				buf.WriteRune(self.next())
				kind = ELL
			}
		default:
			buf.WriteRune(self.ch)
		}

		pos.SetEnd(self.pos, self.row, self.col)
		self.next()
		return Token{
			Pos:    pos,
			Kind:   kind,
			Source: buf.String(),
		}
	}
}
