package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseGlobal() ast.Global {
	switch p.cur.Kind {
	case token.KindEnum.Let:
		return p.parseLet()
	default:
		panic("unreachable")
	}
}
