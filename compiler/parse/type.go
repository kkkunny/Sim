package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	return p.parseSuffixType(p.parsePrimaryType())
}

func (p *Parser) parsePrimaryType() ast.Type {
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
			if len(elems) == 1 {
				return elems[0]
			}
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

func (p *Parser) parseSuffixType(prev ast.Type) ast.Type {
	for {
		switch p.nextToken.Kind {
		case token.KindEnum.Or:
			p.expect(token.KindEnum.Or)
			next := p.parsePrimaryType()
			if ut, ok := prev.(*ast.UnionType); ok {
				prev = &ast.UnionType{Elems: append(ut.Elems, next)}
			} else {
				prev = &ast.UnionType{Elems: []ast.Type{prev, next}}
			}
		default:
			return prev
		}
	}
}
