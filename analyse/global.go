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
	f := &FuncDef{
		Name: node.Name.Source(),
		Ret:  self.analyseOptionType(node.Ret),
	}

	self.localScope = _NewFuncScope(self.pkgScope, f.Ret)
	defer func() {
		self.localScope = nil
	}()

	f.Body = self.analyseBlock(node.Body)
	return f
}
