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
	if !self.pkgScope.SetValue(f.Name, f) {
		// TODO: 编译时异常：变量名冲突
		panic("unreachable")
	}

	self.localScope = _NewFuncScope(self.pkgScope, f.Ret)
	defer func() {
		self.localScope = nil
	}()

	body, end := self.analyseBlock(node.Body)
	f.Body = body
	if !end {
		// TODO: 编译时异常：缺少函数返回值
		panic("unreachable")
	}
	return f
}
