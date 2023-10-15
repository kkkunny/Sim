package parse

import (
	. "github.com/kkkunny/Sim/ast"
)

func (self *Parser) parseTypeOrNil() *Type {
	switch self.nextTok.Kind {
	default:
		return nil
	}
}

func (self *Parser) parseType() Type {
	t := self.parseTypeOrNil()
	if t == nil {
		// TOOD: 报错
		panic("unreachable")
	}
	return *t
}
