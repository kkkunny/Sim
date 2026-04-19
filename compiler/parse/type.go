package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	switch p.nextToken.Kind {
	case token.KindEnum.Lpa:
		begin := p.expect(token.KindEnum.Lpa).Position
		p.skip(token.KindEnum.Br)
		var elems []ast.Type
		for p.nextToken.Kind != token.KindEnum.Rpa {
			elems = append(elems, p.parseType())
			if !p.ifSkip(token.KindEnum.Comma) {
				break
			}
			p.skip(token.KindEnum.Br)
		}
		end := p.expect(token.KindEnum.Rpa).Position

		if !p.ifSkip(token.KindEnum.Arrow) {
			return &ast.TupleType{BeginPosition: begin, Elems: elems, EndPosition: end}
		}

		returnType := optional.Some(p.parseType())
		return &ast.FuncType{BeginPosition: begin, Params: elems, ReturnType: returnType, EndPosition: p.curToken.Position}
	case token.KindEnum.Lba:
		begin := p.expect(token.KindEnum.Lba).Position
		size := p.expect(token.KindEnum.Integer)
		p.expect(token.KindEnum.Rba)
		elem := p.parseType()
		return &ast.ArrayType{BeginPosition: begin, Size: size, Elem: elem}
	default:
		name := p.expect(token.KindEnum.Ident)
		return &ast.IdentType{Name: name}
	}
}
