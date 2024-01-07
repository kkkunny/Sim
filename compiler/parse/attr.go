package parse

import (
	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

func (self *Parser) parseAttrList() (attrs []ast.Attr) {
	for self.skipNextIs(token.AT) {
		begin := self.curTok.Position
		attrname := self.expectNextIs(token.IDENT)
		switch attrname.Source() {
		case "extern":
			attrs = append(attrs, self.parseExtern(begin))
		case "noreturn":
			attrs = append(attrs, self.parseNoReturn(begin))
		case "inline":
			attrs = append(attrs, self.parseInline(begin))
		case "noinline":
			attrs = append(attrs, self.parseNoInline(begin))
		case "var_arg":
			attrs = append(attrs, self.parseVarArg(begin))
		default:
			errors.ThrowIllegalAttr(self.nextTok.Position)
			panic("unreachable")
		}
		self.expectNextIs(token.SEM)
	}
	return attrs
}

func (self *Parser) parseExtern(begin reader.Position) *ast.Extern {
	self.expectNextIs(token.LPA)
	name := self.expectNextIs(token.STRING)
	end := self.expectNextIs(token.RPA).Position
	return &ast.Extern{
		Begin: begin,
		Name:  name,
		End:   end,
	}
}

func (self *Parser) parseNoReturn(begin reader.Position) *ast.NoReturn {
	return &ast.NoReturn{
		Begin: begin,
		End:   self.curTok.Position,
	}
}

func (self *Parser) parseInline(begin reader.Position) *ast.Inline {
	return &ast.Inline{
		Begin: begin,
		End:   self.curTok.Position,
	}
}

func (self *Parser) parseNoInline(begin reader.Position) *ast.NoInline {
	return &ast.NoInline{
		Begin: begin,
		End:   self.curTok.Position,
	}
}

func (self *Parser) parseVarArg(begin reader.Position) *ast.VarArg {
	return &ast.VarArg{
		Begin: begin,
		End:   self.curTok.Position,
	}
}
