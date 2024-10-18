package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/compiler/oldhir"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/reader"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
	"github.com/kkkunny/Sim/compiler/util"
)

func (self *Analyser) analyseImport(node *ast.Import) linkedlist.LinkedList[oldhir.Global] {
	// 包名
	var pkgName string
	var importAll bool
	if alias, ok := node.Alias.Value(); ok && alias.Is(token.IDENT) {
		pkgName = alias.Source()
	} else {
		importAll = alias.Is(token.MUL)
		pkgName = stlslices.Last(node.Paths).Source()
	}

	// 包地址
	paths := stlslices.Map(node.Paths, func(_ int, v token.Token) string {
		return v.Source()
	})
	pkg, err := oldhir.OfficialPackage.GetSon(paths...)
	if err != nil {
		errors.ThrowInvalidPackage(reader.MixPosition(stlslices.Last(node.Paths).Position, stlslices.Last(node.Paths).Position), node.Paths)
	}

	hirs, importErrKind := self.importPackage(pkg, pkgName, importAll)
	if importErrKind != importPackageErrorNone {
		switch importErrKind {
		case importPackageErrorCircular:
			errors.ThrowCircularReference(stlslices.Last(node.Paths).Position, stlslices.Last(node.Paths))
		case importPackageErrorDuplication:
			errors.ThrowIdentifierDuplicationError(stlslices.Last(node.Paths).Position, stlslices.Last(node.Paths))
		case importPackageErrorInvalid:
			errors.ThrowInvalidPackage(reader.MixPosition(stlslices.First(node.Paths).Position, stlslices.Last(node.Paths).Position), node.Paths)
		default:
			panic("unreachable")
		}
	}
	return hirs
}

func (self *Analyser) declTypeDef(node *ast.TypeDef) {
	st := &oldhir.TypeDef{
		Pkg:     self.pkgScope.pkg,
		Public:  node.Public,
		Name:    node.Name.Source(),
		Methods: hashmap.StdWith[string, *oldhir.MethodDef](),
	}

	if !self.pkgScope.SetTypeDef(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
	tad := &oldhir.TypeAliasDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
	}
	if !self.pkgScope.SetTypeDef(tad) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declTrait(node *ast.Trait) {
	trait := &oldhir.Trait{
		Pkg:     self.pkgScope.pkg,
		Public:  node.Public,
		Name:    node.Name.Source(),
		Methods: linkedhashmap.StdWithCap[string, *oldhir.FuncDecl](uint(len(node.Methods))),
	}
	if !self.pkgScope.SetTrait(trait) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) defTypeDef(node *ast.TypeDef) *oldhir.TypeDef {
	gt, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	td := gt.(*oldhir.TypeDef)

	defer self.setSelfType(td)()

	td.Target = self.analyseType(node.Target)
	return td
}

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) *oldhir.TypeAliasDef {
	gt, ok := self.pkgScope.getLocalTypeDef(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	tad := gt.(*oldhir.TypeAliasDef)

	tad.Target = self.analyseType(node.Target)
	return tad
}

func (self *Analyser) defTrait(node *ast.Trait) *oldhir.Trait {
	trait, ok := self.pkgScope.getLocalTrait(node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	self.selfCanBeNil = true
	defer func() {
		self.selfCanBeNil = false
	}()

	for _, methodNode := range node.Methods {
		method := stlval.Ptr(self.analyseFuncDecl(*methodNode))
		if trait.Methods.Contain(method.Name) {
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
	f := &oldhir.FuncDef{
		Pkg:    self.pkgScope.pkg,
		Public: node.Public,
	}
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			f.ExternName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
		case *ast.Inline:
			f.InlineControl = optional.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = optional.Some[bool](false)
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

	if f.Name == "main" && !f.GetType().EqualTo(&oldhir.FuncType{Ret: oldhir.NoThing}) {
		errors.ThrowTypeMismatchError(node.Position(), f.GetType(), &oldhir.FuncType{Ret: oldhir.NoThing})
	}
	if !self.pkgScope.SetValue(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declMethodDef(node *ast.FuncDef) {
	f := &oldhir.MethodDef{
		FuncDef: oldhir.FuncDef{
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
			f.InlineControl = optional.Some[bool](true)
		case *ast.NoInline:
			f.InlineControl = optional.Some[bool](false)
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
	if !ok || !stlval.Is[*oldhir.TypeDef](td) {
		errors.ThrowUnknownIdentifierError(node.SelfType.MustValue().Position, node.SelfType.MustValue())
	}
	f.Scope = td.(*oldhir.TypeDef)
	defer self.setSelfType(f.Scope)()

	f.FuncDecl = self.analyseFuncDecl(node.FuncDecl)

	if f.Scope.Methods.Contain(f.Name) || (oldhir.IsType[*oldhir.StructType](f.Scope.Target) && oldhir.AsType[*oldhir.StructType](f.Scope.Target).Fields.Contain(f.Name)) {
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

	v := &oldhir.GlobalVarDef{
		VarDecl: oldhir.VarDecl{
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

func (self *Analyser) analyseGlobalDef(node ast.Global) oldhir.Global {
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

func (self *Analyser) defFuncDef(node *ast.FuncDef) *oldhir.FuncDef {
	value, ok := self.pkgScope.getLocalValue(node.Name.Source())
	if !ok {
		panic("unreachable")
	}
	f := value.(*oldhir.FuncDef)

	if node.Body.IsNone() {
		return f
	}

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()

	for i, p := range f.Params {
		if p.Name.IsSome() && !self.localScope.SetValue(p.Name.MustValue(), p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.MustValue().Position, node.Params[i].Name.MustValue())
		}
	}

	f.Body = optional.Some(self.analyseFuncBody(node.Body.MustValue()))
	return f
}

func (self *Analyser) defMethodDef(node *ast.FuncDef) *oldhir.MethodDef {
	td, _ := self.pkgScope.getLocalTypeDef(node.SelfType.MustValue().Source())
	st := td.(*oldhir.TypeDef)
	f := st.Methods.Get(node.Name.Source())

	if node.Body.IsNone() {
		return f
	}

	self.localScope = _NewFuncScope(self.pkgScope, f)
	defer func() {
		self.localScope = nil
	}()
	defer self.setSelfType(st)()

	for i, p := range f.Params {
		if p.Name.IsSome() && !self.localScope.SetValue(p.Name.MustValue(), p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.MustValue().Position, node.Params[i].Name.MustValue())
		}
	}

	f.Body = optional.Some(self.analyseFuncBody(node.Body.MustValue()))
	return f
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *oldhir.GlobalVarDef {
	value, ok := self.pkgScope.GetValue("", node.Var.Name.Source())
	if !ok {
		panic("unreachable")
	}
	v := value.(*oldhir.GlobalVarDef)

	if valueNode, ok := node.Value.Value(); ok {
		v.Value = optional.Some(self.expectExpr(v.Type, valueNode))
	} else if v.ExternName == "" {
		v.Value = optional.Some[oldhir.Expr](self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), v.Type))
	}
	return v
}

func (self *Analyser) defMultiGlobalVariable(node *ast.MultipleVariableDef) *oldhir.MultiGlobalVarDef {
	vars := lo.Map(node.Vars, func(item ast.VarDef, _ int) *oldhir.GlobalVarDef {
		value, ok := self.pkgScope.GetValue("", item.Name.Source())
		if !ok {
			panic("unreachable")
		}
		return value.(*oldhir.GlobalVarDef)
	})
	varTypes := lo.Map(vars, func(item *oldhir.GlobalVarDef, _ int) oldhir.Type {
		return item.GetType()
	})

	var value oldhir.Expr
	if valueNode, ok := node.Value.Value(); ok {
		value = self.expectExpr(&oldhir.TupleType{Elems: varTypes}, valueNode)
	} else {
		tupleValue := &oldhir.Tuple{Elems: make([]oldhir.Expr, len(vars))}
		for i, varDef := range node.Vars {
			tupleValue.Elems[i] = self.getTypeDefaultValue(varDef.Type.MustValue().Position(), varTypes[i])
		}
		value = tupleValue
	}
	return &oldhir.MultiGlobalVarDef{
		Vars:  vars,
		Value: value,
	}
}
