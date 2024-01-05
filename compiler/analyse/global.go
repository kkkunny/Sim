package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/hashset"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
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
	if err != nil {
		errors.ThrowInvalidPackage(reader.MixPosition(node.Paths.Front().Position, node.Paths.Back().Position), node.Paths)
	}

	hirs, importErrKind := self.importPackage(pkg, pkgName, importAll)
	if importErrKind != importPackageErrorNone {
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
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Name.Source(),
	}

	if !self.pkgScope.SetTypeDef(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
	tad := &hir.TypeAliasDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
	}
	if !self.pkgScope.SetTypeDef(tad) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declGenericStructDef(node *ast.StructDef) {
	st := &hir.GenericStructDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Name.Source(),
	}

	for _, p := range node.Name.Params.MustValue().Data {
		name := p.Source()
		if st.GenericParams.ContainKey(name) {
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		pt := &hir.GenericIdentType{
			Belong: st,
			Name:   name,
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
	td, ok := self.pkgScope.getLocalTypeDef(node.Name.Name.Source())
	if !ok {
		panic("unreachable")
	}
	st := td.(*hir.StructDef)

	for _, f := range node.Fields {
		st.Fields.Set(f.Name.Source(), hir.Field{
			Public:  f.Public,
			Mutable: f.Mutable,
			Type:    self.analyseType(f.Type),
		})
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

func (self *Analyser) defGenericStructDef(node *ast.StructDef) *hir.GenericStructDef {
	st, ok := self.pkgScope.getLocalGenericStructDef(node.Name.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for iter := st.GenericParams.Iterator(); iter.Next(); {
		self.genericIdentMap.Set(iter.Value().First, iter.Value().Second)
	}
	defer func() {
		self.genericIdentMap.Clear()
	}()

	for _, f := range node.Fields {
		st.Fields.Set(f.Name.Source(), hir.Field{
			Public:  f.Public,
			Mutable: f.Mutable,
			Type:    self.analyseType(f.Type),
		})
	}
	return st
}

func (self *Analyser) analyseGlobalDecl(node ast.Global) {
	switch global := node.(type) {
	case *ast.FuncDef:
		if global.Name.Params.IsNone() {
			self.declFuncDef(global)
		} else {
			self.declGenericFuncDef(global)
		}
	case *ast.MethodDef:
		if global.SelfType.Params.IsNone() {
			if global.Name.Params.IsNone() {
				self.declMethodDef(global)
			} else {
				self.declGenericMethodDef(global)
			}
		} else {
			self.declGenericStructMethodDef(global)
		}
	case *ast.SingleVariableDef:
		self.declSingleGlobalVariable(global)
	case *ast.MultipleVariableDef:
		self.declMultiGlobalVariable(global)
	case *ast.StructDef, *ast.Import, *ast.TypeAlias:
	default:
		panic("unreachable")
	}
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) {
	f := &hir.FuncDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Name.Source(),
		Body:   util.None[*hir.Block](),
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
		errors.ThrowExpectAttribute(node.Name.Position(), new(ast.Extern))
	}

	paramNameSet := hashset.NewHashSet[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			VarDecl: hir.VarDecl{
				Mut:  paramNode.Mutable,
				Type: pt,
				Name: pn,
			},
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Name == "main" && !f.GetType().EqualTo(&hir.FuncType{Ret: hir.Empty}) {
		errors.ThrowTypeMismatchError(node.Name.Position(), f.GetType(), &hir.FuncType{Ret: hir.Empty})
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
}

func (self *Analyser) declMethodDef(node *ast.MethodDef) {
	f := &hir.MethodDef{
		FuncDef: hir.FuncDef{
			Pkg:    self.pkgScope.pkg,
			Public: node.Public,
			Name:   node.Name.Name.Source(),
		},
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

	td, ok := self.pkgScope.getLocalTypeDef(node.SelfType.Name.Source())
	if !ok || !stlbasic.Is[*hir.StructDef](td) {
		errors.ThrowUnknownIdentifierError(node.SelfType.Name.Position, node.SelfType.Name)
	}
	f.Scope = td.(*hir.StructDef)

	defer self.setSelfType(f.Scope)()

	paramNameSet := hashset.NewHashSetWith[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			VarDecl: hir.VarDecl{
				Mut:  paramNode.Mutable,
				Type: pt,
				Name: pn,
			},
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Scope.Fields.ContainKey(f.Name) || f.Scope.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position(), node.Name.Name)
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

	v := &hir.GlobalVarDef{
		VarDecl: hir.VarDecl{
			Mut:  node.Var.Mutable,
			Type: self.analyseType(node.Var.Type.MustValue()),
			Name: node.Var.Name.Source(),
		},
		Pkg:        self.pkgScope.pkg,
		Public:     node.Public,
		ExternName: externName,
	}
	if !self.pkgScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}
}

func (self *Analyser) declMultiGlobalVariable(node *ast.MultipleVariableDef) {
	for _, single := range node.ToSingleList() {
		self.declSingleGlobalVariable(single)
	}
}

func (self *Analyser) declGenericFuncDef(node *ast.FuncDef) {
	f := &hir.GenericFuncDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Name.Source(),
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

	for _, p := range node.Name.Params.MustValue().Data {
		pn := p.Source()
		if f.GenericParams.ContainKey(pn) {
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		pt := &hir.GenericIdentType{
			Belong: f,
			Name:   pn,
		}
		f.GenericParams.Set(pn, pt)
		self.genericIdentMap.Set(pn, pt)
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
			VarDecl: hir.VarDecl{
				Mut:  paramNode.Mutable,
				Type: pt,
				Name: pn,
			},
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)

	if !self.pkgScope.SetGenericFuncDef(f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
}

func (self *Analyser) declGenericStructMethodDef(node *ast.MethodDef) {
	f := &hir.GenericStructMethodDef{
		GenericFuncDef: hir.GenericFuncDef{
			Pkg:    self.pkgScope.pkg,
			Public: node.Public,
			Name:   node.Name.Name.Source(),
		},
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
	f.Scope, ok = self.pkgScope.getLocalGenericStructDef(node.SelfType.Name.Source())
	if !ok {
		errors.ThrowUnknownIdentifierError(node.SelfType.Name.Position, node.SelfType.Name)
	} else if f.Scope.GenericParams.Length() != uint(len(node.SelfType.Params.MustValue().Data)) {
		errors.ThrowParameterNumberNotMatchError(node.SelfType.Position(), f.Scope.GenericParams.Length(), uint(len(node.SelfType.Params.MustValue().Data)))
	}

	scopeGenericParams := f.Scope.GenericParams.Values().ToSlice()
	scopeGenericParamSet := hashset.NewHashSetWithCapacity[string](uint(len(scopeGenericParams)))
	for i, p := range node.SelfType.Params.MustValue().Data {
		pn := p.Source()
		if !scopeGenericParamSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		self.genericIdentMap.Set(pn, scopeGenericParams[i])
	}
	if genericParams, ok := node.Name.Params.Value(); ok {
		for _, p := range genericParams.Data {
			pn := p.Source()
			if scopeGenericParamSet.Contain(pn) || f.GenericParams.ContainKey(pn) {
				errors.ThrowIdentifierDuplicationError(p.Position, p)
			}
			pt := &hir.GenericIdentType{
				Belong: f,
				Name:   pn,
			}
			f.GenericParams.Set(pn, pt)
			self.genericIdentMap.Set(pn, pt)
		}
	}
	defer self.genericIdentMap.Clear()
	defer self.setSelfType(f.GetSelfType())()

	paramNameSet := hashset.NewHashSetWith[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			VarDecl: hir.VarDecl{
				Mut:  paramNode.Mutable,
				Type: pt,
				Name: pn,
			},
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Scope.Fields.ContainKey(f.Name) || f.Scope.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
	f.Scope.Methods.Set(f.Name, f)
}

func (self *Analyser) declGenericMethodDef(node *ast.MethodDef) {
	f := &hir.GenericMethodDef{
		GenericFuncDef: hir.GenericFuncDef{
			Pkg:    self.pkgScope.pkg,
			Public: node.Public,
			Name:   node.Name.Name.Source(),
		},
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
	td, ok := self.pkgScope.getLocalTypeDef(node.SelfType.Name.Source())
	if !ok || !stlbasic.Is[*hir.StructDef](td) {
		errors.ThrowUnknownIdentifierError(node.SelfType.Position(), node.SelfType.Name)
	}
	f.Scope = td.(*hir.StructDef)

	for _, p := range node.Name.Params.MustValue().Data {
		pn := p.Source()
		if f.GenericParams.ContainKey(pn) {
			errors.ThrowIdentifierDuplicationError(p.Position, p)
		}
		pt := &hir.GenericIdentType{
			Belong: f,
			Name:   pn,
		}
		f.GenericParams.Set(pn, pt)
		self.genericIdentMap.Set(pn, pt)
	}
	defer self.genericIdentMap.Clear()
	defer self.setSelfType(f.Scope)()

	paramNameSet := hashset.NewHashSetWith[string]()
	f.Params = lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
		pn := paramNode.Name.Source()
		if !paramNameSet.Add(pn) {
			errors.ThrowIdentifierDuplicationError(paramNode.Name.Position, paramNode.Name)
		}
		pt := self.analyseType(paramNode.Type)
		return &hir.Param{
			VarDecl: hir.VarDecl{
				Mut:  paramNode.Mutable,
				Type: pt,
				Name: pn,
			},
		}
	})
	f.Ret = self.analyseOptionType(node.Ret)
	if f.Scope.Fields.ContainKey(f.Name) || f.Scope.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Name.Position, node.Name.Name)
	}
	f.Scope.Methods.Set(f.Name, f)
}

func (self *Analyser) analyseGlobalDef(node ast.Global) hir.Global {
	switch global := node.(type) {
	case *ast.FuncDef:
		if global.Name.Params.IsNone() {
			return self.defFuncDef(global)
		} else {
			return self.defGenericFuncDef(global)
		}
	case *ast.MethodDef:
		if global.SelfType.Params.IsNone() {
			if global.Name.Params.IsNone() {
				return self.defMethodDef(global)
			} else {
				return self.defGenericMethodDef(global)
			}
		} else {
			return self.defGenericStructMethodDef(global)
		}
	case *ast.SingleVariableDef:
		return self.defSingleGlobalVariable(global)
	case *ast.MultipleVariableDef:
		return self.defMultiGlobalVariable(global)
	case *ast.StructDef, *ast.Import, *ast.TypeAlias:
		return nil
	default:
		panic("unreachable")
	}
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *hir.FuncDef {
	value, ok := self.pkgScope.getLocalValue(node.Name.Name.Source())
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

	f.Body = util.Some(self.analyseFuncBody(node.Body.MustValue()))
	return f
}

func (self *Analyser) defMethodDef(node *ast.MethodDef) *hir.MethodDef {
	td, _ := self.pkgScope.getLocalTypeDef(node.SelfType.Name.Source())
	st := td.(*hir.StructDef)
	f := st.Methods.Get(node.Name.Name.Source()).(*hir.MethodDef)

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()
	defer self.setSelfType(st)()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	f.Body = util.Some(self.analyseFuncBody(node.Body))
	return f
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *hir.GlobalVarDef {
	value, ok := self.pkgScope.GetValue("", node.Var.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*hir.GlobalVarDef)

	if valueNode, ok := node.Value.Value(); ok {
		v.Value = self.expectExpr(v.Type, valueNode)
	} else {
		v.Value = self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), v.Type)
	}
	return v
}

func (self *Analyser) defMultiGlobalVariable(node *ast.MultipleVariableDef) *hir.MultiGlobalVarDef {
	vars := lo.Map(node.Vars, func(item ast.VarDef, _ int) *hir.GlobalVarDef {
		value, ok := self.pkgScope.GetValue("", item.Name.Source())
		if !ok {
			panic("unreachable")
		}
		return value.(*hir.GlobalVarDef)
	})
	varTypes := lo.Map(vars, func(item *hir.GlobalVarDef, _ int) hir.Type {
		return item.GetType()
	})

	var value hir.Expr
	if valueNode, ok := node.Value.Value(); ok {
		value = self.expectExpr(&hir.TupleType{Elems: varTypes}, valueNode)
	} else {
		tupleValue := &hir.Tuple{Elems: make([]hir.Expr, len(vars))}
		for i, varDef := range node.Vars {
			tupleValue.Elems[i] = self.getTypeDefaultValue(varDef.Type.MustValue().Position(), varTypes[i])
		}
		value = tupleValue
	}
	return &hir.MultiGlobalVarDef{
		Vars:  vars,
		Value: value,
	}
}

func (self *Analyser) defGenericFuncDef(node *ast.FuncDef) *hir.GenericFuncDef {
	f, ok := self.pkgScope.getLocalGenericFuncDef(node.Name.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for iter := f.GenericParams.Iterator(); iter.Next(); {
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

	f.Body = self.analyseFuncBody(node.Body.MustValue())
	return f
}

func (self *Analyser) defGenericStructMethodDef(node *ast.MethodDef) *hir.GenericStructMethodDef {
	st, _ := self.pkgScope.getLocalGenericStructDef(node.SelfType.Name.Source())
	f := st.Methods.Get(node.Name.Name.Source())

	for i, iter := 0, f.Scope.GenericParams.Iterator(); iter.Next(); i++ {
		self.genericIdentMap.Set(node.SelfType.Params.MustValue().Data[i].Source(), iter.Value().Second)
	}
	for i, iter := 0, f.GenericParams.Iterator(); iter.Next(); i++ {
		self.genericIdentMap.Set(node.Name.Params.MustValue().Data[i].Source(), iter.Value().Second)
	}
	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.genericIdentMap.Clear()
		self.localScope = nil
	}()
	defer self.setSelfType(f.GetSelfType())()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	f.Body = self.analyseFuncBody(node.Body)
	return f
}

func (self *Analyser) defGenericMethodDef(node *ast.MethodDef) *hir.GenericMethodDef {
	td, _ := self.pkgScope.getLocalTypeDef(node.SelfType.Name.Source())
	st := td.(*hir.StructDef)
	f := st.Methods.Get(node.Name.Name.Source()).(*hir.GenericMethodDef)

	for i, iter := 0, f.GenericParams.Iterator(); iter.Next(); i++ {
		self.genericIdentMap.Set(node.Name.Params.MustValue().Data[i].Source(), iter.Value().Second)
	}
	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.genericIdentMap.Clear()
		self.localScope = nil
	}()
	defer self.setSelfType(f.Scope)()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	f.Body = self.analyseFuncBody(node.Body)
	return f
}
