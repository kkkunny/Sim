package parse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
)

// Attr 属性
type Attr interface {
	Ast
	Attr()
}

// AttrExtern @extern
type AttrExtern struct {
	Pos  utils.Position
	Name lex.Token
}

func NewAttrExtern(pos utils.Position, name lex.Token) *AttrExtern {
	return &AttrExtern{
		Pos:  pos,
		Name: name,
	}
}

func (self AttrExtern) Position() utils.Position {
	return self.Pos
}

func (self AttrExtern) Attr() {}

// AttrLink @link
type AttrLink struct {
	Pos  utils.Position
	Asms []*String
	Libs []*String
}

func NewAttrLink(pos utils.Position, asms, libs []*String) *AttrLink {
	return &AttrLink{
		Pos:  pos,
		Asms: asms,
		Libs: libs,
	}
}

func (self AttrLink) Position() utils.Position {
	return self.Pos
}

func (self AttrLink) Attr() {}

// AttrNoReturn @noreturn
type AttrNoReturn struct {
	Pos utils.Position
}

func NewAttrNoReturn(pos utils.Position) *AttrNoReturn {
	return &AttrNoReturn{Pos: pos}
}

func (self AttrNoReturn) Position() utils.Position {
	return self.Pos
}

func (self AttrNoReturn) Attr() {}

// AttrInline @inline
type AttrInline struct {
	Pos   utils.Position
	Value lex.Token
}

func NewAttrInline(pos utils.Position, v lex.Token) *AttrInline {
	return &AttrInline{
		Pos:   pos,
		Value: v,
	}
}

func (self AttrInline) Position() utils.Position {
	return self.Pos
}

func (self AttrInline) Attr() {}

// AttrInit @init
type AttrInit struct {
	Pos utils.Position
}

func NewAttrInit(pos utils.Position) *AttrInit {
	return &AttrInit{
		Pos: pos,
	}
}

func (self AttrInit) Position() utils.Position {
	return self.Pos
}

func (self AttrInit) Attr() {}

// AttrFini @fini
type AttrFini struct {
	Pos utils.Position
}

func NewAttrFini(pos utils.Position) *AttrFini {
	return &AttrFini{
		Pos: pos,
	}
}

func (self AttrFini) Position() utils.Position {
	return self.Pos
}

func (self AttrFini) Attr() {}

// ****************************************************************

func (self *parser) parseAttr() Attr {
	attrName := self.expectNextIs(lex.Attr)
	switch attrName.Source {
	case "@extern":
		self.expectNextIs(lex.LPA)
		name := self.expectNextIs(lex.IDENT)
		end := self.expectNextIs(lex.RPA).Pos
		return NewAttrExtern(utils.MixPosition(attrName.Pos, end), name)
	case "@link":
		self.expectNextIs(lex.LPA)
		var asms, libs []*String
		for {
			linkname := self.expectNextIs(lex.IDENT)
			switch linkname.Source {
			case "asm":
				self.expectNextIs(lex.ASS)
				asms = append(asms, self.parseStringExpr())
			case "lib":
				self.expectNextIs(lex.ASS)
				libs = append(libs, self.parseStringExpr())
			default:
				self.throwErrorf(linkname.Pos, "unknown link")
			}
			if !self.skipNextIs(lex.COM) {
				break
			}
		}
		end := self.expectNextIs(lex.RPA).Pos
		return NewAttrLink(utils.MixPosition(attrName.Pos, end), asms, libs)
	case "@noreturn":
		return NewAttrNoReturn(attrName.Pos)
	case "@inline":
		self.expectNextIs(lex.LPA)
		var v lex.Token
		if self.skipNextIs(lex.FALSE) {
			v = self.curTok
		} else {
			v = self.expectNextIs(lex.TRUE)
		}
		end := self.expectNextIs(lex.RPA).Pos
		return NewAttrInline(utils.MixPosition(attrName.Pos, end), v)
	case "@init":
		return NewAttrInit(attrName.Pos)
	case "@fini":
		return NewAttrFini(attrName.Pos)
	default:
		self.throwErrorf(attrName.Pos, "unknown attribute")
		return nil
	}
}
