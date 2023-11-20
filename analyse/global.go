package analyse

import (
	"path/filepath"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/config"
	errors "github.com/kkkunny/Sim/error"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseImport(node *ast.Import) linkedlist.LinkedList[Global] {
	// 包名
	var pkgName string
	if alias, ok := node.Alias.Value(); ok && alias.Is(token.IDENT) {
		pkgName = alias.Source()
	} else {
		pkgName = node.Paths.Back().Source()
	}
	if self.pkgScope.externs.ContainKey(pkgName) {
		errors.ThrowIdentifierDuplicationError(node.Paths.Back().Position, node.Paths.Back())
	}

	// 包地址（唯一标识符）
	paths := iterator.Map[token.Token, string, dynarray.DynArray[string]](node.Paths, func(v token.Token) string {
		return v.Source()
	}).ToSlice()
	pkgPath := filepath.Join(append([]string{config.ROOT}, paths...)...)

	// 检查循环导入
	if self.checkLoopImport(pkgPath) {
		errors.ThrowCircularImport(node.Paths.Back().Position, node.Paths.Back())
	}
	// 如果有缓存则直接返回
	if pkgScope := self.pkgs.Get(pkgPath); pkgScope != nil {
		self.pkgScope.externs.Set(pkgName, pkgScope)
		return linkedlist.LinkedList[Global]{}
	}

	// 语义分析目标包
	pkgAsts, err := parse.ParseDir(pkgPath)
	if err != nil {
		errors.ThrowInvalidPackage(reader.MixPosition(node.Paths.Front().Position, node.Paths.Back().Position), node.Paths)
	}
	pkgAnalyser := newSon(self, pkgPath, pkgAsts)
	pkgMeans := pkgAnalyser.Analyse()
	// 放进缓存
	self.pkgs.Set(pkgPath, pkgAnalyser.pkgScope)
	self.pkgScope.externs.Set(pkgName, pkgAnalyser.pkgScope)
	return pkgMeans
}

func (self *Analyser) declTypeDef(node *ast.StructDef) {
	st := &StructDef{
		Public: node.Public,
		Name:   node.Name.Source(),
		Fields: linkedhashmap.NewLinkedHashMap[string, Type](),
	}
	if !self.pkgScope.SetStruct(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) defTypeDef(node *ast.StructDef) *StructDef {
	st, ok := self.pkgScope.GetStruct("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for _, f := range node.Fields {
		fn := f.First.Source()
		ft := self.analyseType(f.Second)
		st.Fields.Set(fn, ft)
	}
	return st
}

func (self *Analyser) analyseGlobalDecl(node ast.Global) {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		self.declFuncDef(globalNode)
	case *ast.Variable:
		self.declGlobalVariable(globalNode)
	case *ast.StructDef, *ast.Import:
	default:
		panic("unreachable")
	}
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) {
	paramNameSet := hashset.NewHashSet[string]()
	params := lo.Map(node.Params, func(paramNode ast.Param, index int) *Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f := &FuncDef{
		Public: node.Public,
		Name:   node.Name.Source(),
		Params: params,
		Ret:    self.analyseOptionType(node.Ret),
	}
	if f.Name == "main" && !f.Ret.Equal(U8) {
		pos := stlbasic.TernaryAction(node.Ret.IsNone(), func() reader.Position {
			return node.Name.Position
		}, func() reader.Position {
			ret, _ := node.Ret.Value()
			return ret.Position()
		})
		errors.ThrowTypeMismatchError(pos, f.Ret, U8)
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declGlobalVariable(node *ast.Variable) {
	v := &Variable{
		Public: node.Public,
		Mut:    node.Mutable,
		Type:   self.analyseType(node.Type),
		Name:   node.Name.Source(),
	}
	if !self.pkgScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) analyseGlobalDef(node ast.Global) Global {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		return self.defFuncDef(globalNode)
	case *ast.StructDef:
		return self.defTypeDef(globalNode)
	case *ast.Variable:
		return self.defGlobalVariable(globalNode)
	case *ast.Import:
		return nil
	default:
		panic("unreachable")
	}
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *FuncDef {
	value, ok := self.pkgScope.GetValue("", node.Name.Source())
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
			errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
		}
	}

	body, jump := self.analyseBlock(node.Body, nil)
	f.Body = body
	if jump != BlockEofReturn {
		if !f.Ret.Equal(Empty) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		f.Body.Stmts.PushBack(&Return{Value: util.None[Expr]()})
	}
	return f
}

func (self *Analyser) defGlobalVariable(node *ast.Variable) *Variable {
	value, ok := self.pkgScope.GetValue("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*Variable)

	v.Value = self.expectExpr(v.Type, node.Value)
	return v
}
