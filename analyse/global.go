package analyse

import (
	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
)

func (self *Analyser) analyseGlobal(node ast.Global) Global {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		return self.analyseFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFuncDef(node *ast.FuncDef) *FuncDef {
	body := self.analyseBlock(node.Body)
	return &FuncDef{
		Name: node.Name.Source(),
		Body: body,
	}
}
