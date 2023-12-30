package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/hashset"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseImport(node *ast.Import) linkedlist.LinkedList[hir.Global] {
	// 包名
	var pkgName string
	var importAll bool
	if alias, ok := node.Alias.Value(); ok && alias.Is(token.IDENT) {
		pkgName = alias.Source()
	} else {
		importAll = alias.Is(token.MUL)
		pkgName = node.Paths.Back().Source()
	}

	// 包地址
	paths := stliter.Map[token.Token, string, dynarray.DynArray[string]](node.Paths, func(v token.Token) string {
		return v.Source()
	}).ToSlice()
	pkg, err := hir.OfficialPackage.GetSon(paths...)
	if err != nil{
		errors.ThrowInvalidPackage(reader.MixPosition(node.Paths.Front().Position, node.Paths.Back().Position), node.Paths)
	}

	hirs, importErrKind := self.importPackage(pkg, pkgName, importAll)
	if importErrKind != importPackageErrorNone{
		switch importErrKind {
		case importPackageErrorCircular:
			errors.ThrowCircularReference(node.Paths.Back().Position, node.Paths.Back())
		case importPackageErrorDuplication:
			errors.ThrowIdentifierDuplicationError(node.Paths.Back().Position, node.Paths.Back())
		case importPackageErrorInvalid:
			errors.ThrowInvalidPackage(reader.MixPosition(node.Paths.Front().Position, node.Paths.Back().Position), node.Paths)
		default:
			panic("unreachable")
		}
	}
	return hirs
}

func (self *Analyser) declStructDef(node *ast.StructDef) {
	st := &hir.StructDef{
		Pkg: self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
		Fields: linkedhashmap.NewLinkedHashMap[string, pair.Pair[bool, hir.Type]](),
	}
	if !self.pkgScope.SetTypeDef(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
	tad := &hir.TypeAliasDef{
		Pkg: self.pkgScope.pkg,
		Public: node.Public,
		Name: node.Name.Source(),
	}
	if !self.pkgScope.SetTypeDef(tad) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declGenericStructDef(node *ast.GenericStructDef) {
	st := &hir.GenericStructDef{
		Pkg: self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Name.Source(),
		Fields: linkedhashmap.NewLinkedHashMap[string, pair.Pair[bool, hir.Type]](),
	}

	for _, p := range node.Name.Params{
		name := p.Source()
		if st.GenericParams.ContainKey(name){
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		pt := &hir.GenericIdentType{
			Belong: st,
			Name: name,
		}
		st.GenericParams.Set(name, pt)
		self.genericIdentMap.Set(name, pt)
	}
	defer self.genericIdentMap.Clear()

	if !self.pkgScope.SetGenericStructDef(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
}

func (self *Analyser) defStructDef(node *ast.StructDef) *hir.StructDef {
	td, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	st := td.(*hir.StructDef)

	for _, f := range node.Fields {
		fn := f.B.Source()
		ft := pair.NewPair(f.A, self.analyseType(f.C))
		st.Fields.Set(fn, ft)
	}
	return st
}

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) *hir.TypeAliasDef {
	td, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	tad := td.(*hir.TypeAliasDef)

	tad.Target = self.analyseType(node.Type)
	return tad
}

func (self *Analyser) defGenericStructDef(node *ast.GenericStructDef) *hir.GenericStructDef {
	st, ok := self.pkgScope.getLocalGenericStructDef(node.Name.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for iter:=st.GenericParams.Iterator(); iter.Next(); {
		self.genericIdentMap.Set(iter.Value().First, iter.Value().Second)
	}
	defer func() {
		self.genericIdentMap.Clear()
	}()

	for _, f := range node.Fields {
		fn := f.B.Source()
		ft := pair.NewPair(f.A, self.analyseType(f.C))
		st.Fields.Set(fn, ft)
	}
	return st
}

func (self *Analyser) analyseGlobalDecl(node ast.Global) {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		self.declFuncDef(globalNode)
	case *ast.MethodDef:
		self.declMethodDef(globalNode)
	case *ast.SingleVariableDef:
		self.declSingleGlobalVariable(globalNode)
	case *ast.MultipleVariableDef:
		self.declMultiGlobalVariable(globalNode)
	case *ast.GenericFuncDef:
		self.declGenericFuncDef(globalNode)
	case *ast.GenericStructMethodDef:
		self.declGenericStructMethodDef(globalNode)
	case *ast.StructDef, *ast.Import, *ast.TypeAlias, *ast.GenericStructDef:
	default:
		panic("unreachable")
	}
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) {
	f := &hir.FuncDef{
		Pkg: self.pkgScope.pkg,
		Public:     node.Public,
		Name:       node.Name.Source(),
		Body: util.None[*hir.Block](),
	}
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			f.ExternName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
		case *ast.NoReturn:
			f.NoReturn = true
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		default:
			panic("unreachable")
		}
	}

	if node.Body.IsNone() && f.ExternName == "" {
		errors.ThrowExpectAttribute(node.Name.Position, new(ast.Extern))
	}

	paramNameSet := hashset.NewHashSet[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Name == "main" && !f.Ret.EqualTo(hir.U8) {
		pos := stlbasic.TernaryAction(node.Ret.IsNone(), func() reader.Position {
			return node.Name.Position
		}, func() reader.Position {
			ret, _ := node.Ret.Value()
			return ret.Position()
		})
		errors.ThrowTypeMismatchError(pos, f.Ret, hir.U8)
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declMethodDef(node *ast.MethodDef) {
	f := &hir.MethodDef{
		Pkg: self.pkgScope.pkg,
		Public:    node.Public,
		Name:      node.Name.Source(),
	}
	for _, attrObj := range node.Attrs {
		switch attrObj.(type) {
		case *ast.NoReturn:
			f.NoReturn = true
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		default:
			panic("unreachable")
		}
	}

	td, ok := self.pkgScope.getLocalTypeDef(node.Scope.Source())
	if !ok || !stlbasic.Is[*hir.StructDef](td){
		errors.ThrowUnknownIdentifierError(node.Scope.Position, node.Scope)
	}
	f.Scope = td.(*hir.StructDef)
	f.SelfParam = &hir.Param{
		Mut:  node.ScopeMutable,
		Type: f.Scope,
		Name: token.SELFVALUE.String(),
	}

	defer self.setSelfType(f.Scope)()

	paramNameSet := hashset.NewHashSetWith[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Scope.Fields.ContainKey(f.Name) || f.Scope.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	f.Scope.Methods.Set(f.Name, f)
}

func (self *Analyser) declSingleGlobalVariable(node *ast.SingleVariableDef) {
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

	v := &hir.VarDef{
		Pkg: self.pkgScope.pkg,
		Public:     node.Public,
		Mut:        node.Var.Mutable,
		Type:       self.analyseType(node.Var.Type.MustValue()),
		ExternName: externName,
		Name:       node.Var.Name.Source(),
	}
	if !self.pkgScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}
}

func (self *Analyser) declMultiGlobalVariable(node *ast.MultipleVariableDef) {
	for _, single := range node.ToSingleList(){
		self.declSingleGlobalVariable(single)
	}
}

func (self *Analyser) declGenericFuncDef(node *ast.GenericFuncDef) {
	f := &hir.GenericFuncDef{
		Pkg: self.pkgScope.pkg,
		Public:     node.Public,
		Name:       node.Name.Name.Source(),
	}
	for _, attrObj := range node.Attrs {
		switch attrObj.(type) {
		case *ast.NoReturn:
			f.NoReturn = true
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		default:
			panic("unreachable")
		}
	}

	for _, p := range node.Name.Params{
		name := p.Source()
		if f.GenericParams.ContainKey(name){
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		pt := &hir.GenericIdentType{
			Belong: f,
			Name: name,
		}
		f.GenericParams.Set(name, pt)
		self.genericIdentMap.Set(name, pt)
	}
	defer self.genericIdentMap.Clear()

	paramNameSet := hashset.NewHashSet[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)

	if !self.pkgScope.SetGenericFuncDef(f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
}

func (self *Analyser) declGenericStructMethodDef(node *ast.GenericStructMethodDef) {
	f := &hir.GenericStructMethodDef{
		Pkg: self.pkgScope.pkg,
		Public:    node.Public,
		Name:      node.Name.Source(),
	}
	for _, attrObj := range node.Attrs {
		switch attrObj.(type) {
		case *ast.NoReturn:
			f.NoReturn = true
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		default:
			panic("unreachable")
		}
	}

	var ok bool
	f.Scope, ok = self.pkgScope.getLocalGenericStructDef(node.Scope.Name.Source())
	if !ok{
		errors.ThrowUnknownIdentifierError(node.Scope.Name.Position, node.Scope.Name)
	}else if f.Scope.GenericParams.Length() != uint(len(node.Scope.Params)){
		errors.ThrowParameterNumberNotMatchError(node.Scope.Position(), f.Scope.GenericParams.Length(), uint(len(node.Scope.Params)))
	}

	scopeGenericParams := f.Scope.GenericParams.Values().ToSlice()
	scopeGenericParamSet := hashset.NewHashSetWithCapacity[string](uint(len(scopeGenericParams)))
	for i, p := range node.Scope.Params{
		name := p.Source()
		if !scopeGenericParamSet.Add(name){
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		self.genericIdentMap.Set(name, scopeGenericParams[i])
	}
	defer self.genericIdentMap.Clear()

	f.SelfParam = &hir.Param{
		Mut:  node.ScopeMutable,
		Type: f.GetSelfType(),
		Name: token.SELFVALUE.String(),
	}

	defer self.setSelfType(f.GetSelfType())()

	paramNameSet := hashset.NewHashSetWith[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			Mut:  paramNode.Mutable,
			Type: pt,
			Name: pn,
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Scope.Fields.ContainKey(f.Name) || f.Scope.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	f.Scope.Methods.Set(f.Name, f)
}

func (self *Analyser) analyseGlobalDef(node ast.Global) hir.Global {
	switch globalNode := node.(type) {
	case *ast.FuncDef:
		return self.defFuncDef(globalNode)
	case *ast.MethodDef:
		return self.defMethodDef(globalNode)
	case *ast.SingleVariableDef:
		return self.defSingleGlobalVariable(globalNode)
	case *ast.MultipleVariableDef:
		return self.defMultiGlobalVariable(globalNode)
	case *ast.GenericFuncDef:
		return self.defGenericFuncDef(globalNode)
	case *ast.GenericStructMethodDef:
		return self.defGenericStructMethodDef(globalNode)
	case *ast.StructDef, *ast.Import, *ast.TypeAlias, *ast.GenericStructDef:
		return nil
	default:
		panic("unreachable")
	}
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *hir.FuncDef {
	value, ok := self.pkgScope.getLocalValue(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	f := value.(*hir.FuncDef)

	if node.Body.IsNone() {
		return f
	}

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	body, jump := self.analyseBlock(node.Body.MustValue(), nil)
	f.Body = util.Some(body)
	if jump != hir.BlockEofReturn {
		if !hir.IsEmptyType(f.Ret) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		body.Stmts.PushBack(&hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		})
	}
	return f
}

func (self *Analyser) defMethodDef(node *ast.MethodDef) *hir.MethodDef {
	td, _ := self.pkgScope.getLocalTypeDef(node.Scope.Source())
	st := td.(*hir.StructDef)
	f := st.Methods.Get(node.Name.Source()).(*hir.MethodDef)

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()
	defer self.setSelfType(st)()
	defer self.setSelfValue(f.SelfParam)()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	var jump hir.BlockEof
	f.Body, jump = self.analyseBlock(node.Body, nil)
	if jump != hir.BlockEofReturn {
		if !hir.IsEmptyType(f.Ret) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		f.Body.Stmts.PushBack(&hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		})
	}
	return f
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *hir.VarDef {
	value, ok := self.pkgScope.GetValue("", node.Var.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*hir.VarDef)

	if valueNode, ok := node.Value.Value(); ok{
		v.Value = self.expectExpr(v.Type, valueNode)
	}else{
		v.Value = self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), v.Type)
	}
	return v
}

func (self *Analyser) defMultiGlobalVariable(node *ast.MultipleVariableDef) *hir.MultiVarDef {
	vars := lo.Map(node.Vars, func(item ast.VarDef, _ int) *hir.VarDef {
		value, ok := self.pkgScope.GetValue("", item.Name.Source())
		if !ok {
			panic("unreachable")
		}
		return value.(*hir.VarDef)
	})
	varTypes := lo.Map(vars, func(item *hir.VarDef, _ int) hir.Type {
		return item.GetType()
	})

	var value hir.Expr
	if valueNode, ok := node.Value.Value(); ok{
		value = self.expectExpr(&hir.TupleType{Elems: varTypes}, valueNode)
	}else{
		tupleValue := &hir.Tuple{Elems: make([]hir.Expr, len(vars))}
		for i, varDef := range node.Vars{
			tupleValue.Elems[i] = self.getTypeDefaultValue(varDef.Type.MustValue().Position(), varTypes[i])
		}
		value = tupleValue
	}
	return &hir.MultiVarDef{
		Vars: vars,
		Value: value,
	}
}

func (self *Analyser) defGenericFuncDef(node *ast.GenericFuncDef) *hir.GenericFuncDef {
	f, ok := self.pkgScope.getLocalGenericFuncDef(node.Name.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for iter:=f.GenericParams.Iterator(); iter.Next(); {
		self.genericIdentMap.Set(iter.Value().First, iter.Value().Second)
	}
	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.genericIdentMap.Clear()
		self.localScope = nil
	}()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	var jump hir.BlockEof
	f.Body, jump = self.analyseBlock(node.Body, nil)
	if jump != hir.BlockEofReturn {
		if !hir.IsEmptyType(f.Ret) {
			errors.ThrowMissingReturnValueError(node.Name.Name.Position, f.Ret)
		}
		f.Body.Stmts.PushBack(&hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		})
	}
	return f
}

func (self *Analyser) defGenericStructMethodDef(node *ast.GenericStructMethodDef) *hir.GenericStructMethodDef {
	st, _ := self.pkgScope.getLocalGenericStructDef(node.Scope.Name.Source())
	f := st.Methods.Get(node.Name.Source())

	var i int
	for iter:=f.Scope.GenericParams.Iterator(); iter.Next(); i++{
		self.genericIdentMap.Set(node.Scope.Params[i].Source(), iter.Value().Second)
	}
	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.genericIdentMap.Clear()
		self.localScope = nil
	}()
	defer self.setSelfType(f.GetSelfType())()
	defer self.setSelfValue(f.SelfParam)()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	var jump hir.BlockEof
	f.Body, jump = self.analyseBlock(node.Body, nil)
	if jump != hir.BlockEofReturn {
		if !hir.IsEmptyType(f.Ret) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		f.Body.Stmts.PushBack(&hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		})
	}
	return f
}
