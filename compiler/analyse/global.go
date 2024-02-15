package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
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

func (self *Analyser) declTypeDef(node *ast.TypeDef) {
	st := &hir.TypeDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
	}

	if !self.pkgScope.SetTypeDef(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
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

func (self *Analyser) declTrait(node *ast.Trait) {
	trait := &hir.Trait{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
	}
	if !self.pkgScope.SetTrait(trait) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) defTypeDef(node *ast.TypeDef) *hir.TypeDef {
	gt, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	td := gt.(*hir.TypeDef)

	td.Target = self.analyseType(node.Target)
	return td
}

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) *hir.TypeAliasDef {
	gt, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	tad := gt.(*hir.TypeAliasDef)

	tad.Target = self.analyseType(node.Target)
	return tad
}

func (self *Analyser) defTrait(node *ast.Trait) *hir.Trait {
	trait, ok := self.pkgScope.getLocalTrait(node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for _, methodNode := range node.Methods {
		method := stlbasic.Ptr(self.analyseFuncDecl(*methodNode))
		if trait.Methods.ContainKey(method.Name) {
			errors.ThrowIdentifierDuplicationError(methodNode.Name.Position, methodNode.Name)
		}
		trait.Methods.Set(method.Name, method)
	}
	return trait
}

func (self *Analyser) analyseGlobalDecl(node ast.Global) {
	switch global := node.(type) {
	case *ast.FuncDef:
		if global.SelfType.IsNone() {
			self.declFuncDef(global)
		} else {
			self.declMethodDef(global)
		}
	case *ast.SingleVariableDef:
		self.declSingleGlobalVariable(global)
	case *ast.MultipleVariableDef:
		self.declMultiGlobalVariable(global)
	case *ast.TypeDef, *ast.Import, *ast.TypeAlias, *ast.Trait:
	default:
		panic("unreachable")
	}
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) {
	f := &hir.FuncDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
	}
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			f.ExternName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		case *ast.VarArg:
			f.VarArg = true
		default:
			panic("unreachable")
		}
	}

	if node.Body.IsNone() && f.ExternName == "" {
		errors.ThrowExpectAttribute(node.Position(), new(ast.Extern))
	}

	f.FuncDecl = self.analyseFuncDecl(node.FuncDecl)

	if f.Name == "main" && !f.GetType().EqualTo(&hir.FuncType{Ret: hir.NoThing}) {
		errors.ThrowTypeMismatchError(node.Position(), f.GetType(), &hir.FuncType{Ret: hir.NoThing})
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declMethodDef(node *ast.FuncDef) {
	f := &hir.MethodDef{
		FuncDef: hir.FuncDef{
			Pkg:    self.pkgScope.pkg,
			Public: node.Public,
		},
	}
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			f.ExternName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
		case *ast.Inline:
			f.InlineControl = util.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = util.Some[bool](false)
		case *ast.VarArg:
			f.VarArg = true
		default:
			panic("unreachable")
		}
	}

	if node.Body.IsNone() && f.ExternName == "" {
		errors.ThrowExpectAttribute(node.Position(), new(ast.Extern))
	}

	td, ok := self.pkgScope.getLocalTypeDef(node.SelfType.MustValue().Source())
	if !ok || !stlbasic.Is[*hir.TypeDef](td) {
		errors.ThrowUnknownIdentifierError(node.SelfType.MustValue().Position, node.SelfType.MustValue())
	}
	f.Scope = td.(*hir.TypeDef)
	defer self.setSelfType(f.Scope)()

	f.FuncDecl = self.analyseFuncDecl(node.FuncDecl)

	if f.Scope.Methods.ContainKey(f.Name) || (hir.IsType[*hir.StructType](f.Scope.Target) && hir.AsType[*hir.StructType](f.Scope.Target).Fields.ContainKey(f.Name)) {
		errors.ThrowIdentifierDuplicationError(node.Position(), node.Name)
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

func (self *Analyser) analyseGlobalDef(node ast.Global) hir.Global {
	switch global := node.(type) {
	case *ast.FuncDef:
		if global.SelfType.IsNone() {
			return self.defFuncDef(global)
		} else {
			return self.defMethodDef(global)
		}
	case *ast.SingleVariableDef:
		return self.defSingleGlobalVariable(global)
	case *ast.MultipleVariableDef:
		return self.defMultiGlobalVariable(global)
	case *ast.TypeDef, *ast.Import, *ast.TypeAlias, *ast.Trait:
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

	f.Body = util.Some(self.analyseFuncBody(node.Body.MustValue()))
	return f
}

func (self *Analyser) defMethodDef(node *ast.FuncDef) *hir.MethodDef {
	td, _ := self.pkgScope.getLocalTypeDef(node.SelfType.MustValue().Source())
	st := td.(*hir.TypeDef)
	f := st.Methods.Get(node.Name.Source()).(*hir.MethodDef)

	if node.Body.IsNone() {
		return f
	}

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

	f.Body = util.Some(self.analyseFuncBody(node.Body.MustValue()))
	return f
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *hir.GlobalVarDef {
	value, ok := self.pkgScope.GetValue("", node.Var.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*hir.GlobalVarDef)

	if valueNode, ok := node.Value.Value(); ok {
		v.Value = util.Some(self.expectExpr(v.Type, valueNode))
	} else if v.ExternName == "" {
		v.Value = util.Some[hir.Expr](self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), v.Type))
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
