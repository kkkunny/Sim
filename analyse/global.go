package analyse

import (
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/token"
)

func (self *Analyser) analyseGlobalDecl(node ast.Global) {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		self.declFuncDef(globalNode)
	}
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) {
	paramNameSet := hashset.NewHashSet[string]()
	params := lo.Map(node.Params, func(item pair.Pair[token.Token, ast.Type], index int) *Param {
		name := item.First.Source()
		if !paramNameSet.Push(name) {
			// TODO: 编译时异常：变量名冲突
			panic("unreachable")
		}
		return &Param{
			Type: self.analyseType(item.Second),
			Name: name,
		}
	})
	f := &FuncDef{
		Name:   node.Name.Source(),
		Params: params,
		Ret:    self.analyseOptionType(node.Ret),
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		// TODO: 编译时异常：变量名冲突
		panic("unreachable")
	}
}

func (self *Analyser) analyseGlobalDef(node ast.Global) Global {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		return self.defFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *FuncDef {
	value, ok := self.pkgScope.GetValue(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	f := value.(*FuncDef)

	self.localScope = _NewFuncScope(self.pkgScope, f.Ret)
	defer func() {
		self.localScope = nil
	}()

	for _, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			// TODO: 编译时异常：变量名冲突
			panic("unreachable")
		}
	}

	body, end := self.analyseBlock(node.Body)
	f.Body = body
	if !end {
		// TODO: 编译时异常：缺少函数返回值
		panic("unreachable")
	}
	return f
}
