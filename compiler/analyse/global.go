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

	"github.com/kkkunny/Sim/mean"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/config"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseImport(node *ast.Import) linkedlist.LinkedList[mean.Global] {
	// 包名
	var pkgName string
	var importAll bool
	if alias, ok := node.Alias.Value(); ok && alias.Is(token.IDENT) {
		pkgName = alias.Source()
	} else {
		importAll = alias.Is(token.MUL)
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
		if importAll {
			self.pkgScope.links.Add(pkgScope)
		} else {
			self.pkgScope.externs.Set(pkgName, pkgScope)
		}
		return linkedlist.LinkedList[mean.Global]{}
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
	if importAll {
		self.pkgScope.links.Add(pkgAnalyser.pkgScope)
	} else {
		self.pkgScope.externs.Set(pkgName, pkgAnalyser.pkgScope)
	}
	return pkgMeans
}

// 导入buildin包
func (self *Analyser) importBuildInPackage() linkedlist.LinkedList[mean.Global] {
	dir := util.GetBuildInPackagePath()

	// 如果有缓存则直接返回
	if pkgScope := self.pkgs.Get(dir); pkgScope != nil {
		self.pkgScope.links.Add(pkgScope)
		return linkedlist.LinkedList[mean.Global]{}
	}

	// 语义分析目标包
	pkgAsts, err := parse.ParseDir(dir)
	if err != nil {
		// HACK: 报编译器异常而不是直接panic
		panic(err)
	}
	pkgAnalyser := newSon(self, dir, pkgAsts)
	pkgMeans := pkgAnalyser.Analyse()
	// 放进缓存
	self.pkgs.Set(dir, pkgAnalyser.pkgScope)
	self.pkgScope.links.Add(pkgAnalyser.pkgScope)
	return pkgMeans
}

func (self *Analyser) declTypeDef(node *ast.StructDef) {
	st := &mean.StructDef{
		Public: node.Public,
		Name:   node.Name.Source(),
		Fields: linkedhashmap.NewLinkedHashMap[string, mean.Type](),
	}
	if !self.pkgScope.SetStruct(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
}

func (self *Analyser) defTypeDef(node *ast.StructDef) *mean.StructDef {
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

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) {
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
	var externName string
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			externName = attr.Name.Source()
			externName = util.ParseEscapeCharacter(externName[1:len(externName)-1], `\"`, `"`)
		default:
			panic("unreachable")
		}
	}

	if node.Body.IsNone() && externName == "" {
		errors.ThrowExpectAttribute(node.Name.Position, new(ast.Extern))
	}

	paramNameSet := hashset.NewHashSet[string]()
	params := lo.Map(node.Params, func(paramNode ast.Param, index int) *mean.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &mean.Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f := &mean.FuncDef{
		Public:     node.Public,
		ExternName: externName,
		Name:       node.Name.Source(),
		Params:     params,
		Ret:        self.analyseOptionType(node.Ret),
		Body:       util.None[*mean.Block](),
	}
	if f.Name == "main" && !f.Ret.Equal(mean.U8) {
		pos := stlbasic.TernaryAction(node.Ret.IsNone(), func() reader.Position {
			return node.Name.Position
		}, func() reader.Position {
			ret, _ := node.Ret.Value()
			return ret.Position()
		})
		errors.ThrowTypeMismatchError(pos, f.Ret, mean.U8)
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declGlobalVariable(node *ast.Variable) {
	var externName string
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			externName = attr.Name.Source()
			externName = util.ParseEscapeCharacter(externName[1:len(externName)-1], `\"`, `"`)
		default:
			panic("unreachable")
		}
	}

	v := &mean.Variable{
		Public:     node.Public,
		Mut:        node.Mutable,
		Type:       self.analyseType(node.Type),
		ExternName: externName,
		Name:       node.Name.Source(),
	}
	if !self.pkgScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) analyseGlobalDef(node ast.Global) mean.Global {
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

func (self *Analyser) defFuncDef(node *ast.FuncDef) *mean.FuncDef {
	value, ok := self.pkgScope.GetValue("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	f := value.(*mean.FuncDef)

	if node.Body.IsNone() {
		return f
	}

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()

	for _, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
		}
	}

	body, jump := self.analyseBlock(node.Body.MustValue(), nil)
	f.Body = util.Some(body)
	if jump != mean.BlockEofReturn {
		if !f.Ret.Equal(mean.Empty) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		body.Stmts.PushBack(&mean.Return{
			Func:  f,
			Value: util.None[mean.Expr](),
		})
	}
	return f
}

func (self *Analyser) defGlobalVariable(node *ast.Variable) *mean.Variable {
	value, ok := self.pkgScope.GetValue("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*mean.Variable)

	v.Value = self.expectExpr(v.Type, node.Value)
	return v
}
