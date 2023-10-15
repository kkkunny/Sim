package analyse

import (
	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseType(node ast.Type) Type {
	switch node {
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) Type {
	t, ok := node.Value()
	if !ok {
		return Empty
	}
	return self.analyseType(t)
}
