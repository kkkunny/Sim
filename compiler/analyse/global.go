package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"
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
	paths := iterator.Map[token.Token, string, dynarray.DynArray[string]](node.Paths, func(v token.Token) string {
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
		Methods: hashmap.NewHashMap[string, *hir.MethodDef](),
	}
	if !self.pkgScope.SetStruct(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declGenericStructDef(node *ast.GenericStructDef) {
	genericParams := linkedhashmap.NewLinkedHashMap[string, *hir.GenericParam]()
	for _, param := range node.GenericParams{
		pn := param.First.Source()
		if genericParams.ContainKey(pn){
			errors.ThrowIdentifierDuplicationError(param.First.Position, param.First)
		}
		genericParam := &hir.GenericParam{Name: pn}

		if constraintNode, ok := param.Second.Value(); ok{
			var pkgName string
			if pkgToken, ok := constraintNode.Pkg.Value(); ok {
				pkgName = pkgToken.Source()
				if !self.pkgScope.externs.ContainKey(pkgName) {
					errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
				}
			}
			traitNode, ok := self.pkgScope.GetTrait(pkgName, constraintNode.Name.Source())
			if !ok{
				errors.ThrowUnknownIdentifierError(constraintNode.Name.Position, constraintNode.Name)
			}
			genericParam.Constraint = util.Some(self.defTrait(genericParam, traitNode))
		}
		genericParams.Set(pn, genericParam)
	}

	st := &hir.GenericStructDef{
		Pkg: self.pkgScope.pkg,
		Public: node.Public,
		Name:   node.Name.Source(),
		GenericParams: genericParams,
		Fields: linkedhashmap.NewLinkedHashMap[string, pair.Pair[bool, hir.Type]](),
	}
	if !self.pkgScope.SetGenericStruct(st) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
	if !self.pkgScope.DeclTypeAlias(node.Name.Source(), node){
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) defStructDef(node *ast.StructDef) *hir.StructDef {
	st, ok := self.pkgScope.GetStruct("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for _, f := range node.Fields {
		fn := f.B.Source()
		ft := pair.NewPair(f.A, self.analyseType(f.C))
		st.Fields.Set(fn, ft)
	}
	return st
}

func (self *Analyser) defGenericStructDef(node *ast.GenericStructDef) *hir.GenericStructDef {
	st, ok := self.pkgScope.getLocalGenericStruct(node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	for iter:=st.GenericParams.Iterator(); iter.Next(); {
		if !self.pkgScope.typeAlias.ContainKey(iter.Value().First){
			self.pkgScope.typeAlias.Set(iter.Value().First, pair.NewPair[bool, either.Either[*ast.TypeAlias, hir.Type]](false, either.Right[*ast.TypeAlias, hir.Type](iter.Value().Second)))
			defer func() {
				self.pkgScope.typeAlias.Remove(iter.Value().First)
			}()
		}else{
			back := self.pkgScope.typeAlias.Get(iter.Value().First)
			self.pkgScope.typeAlias.Set(iter.Value().First, pair.NewPair[bool, either.Either[*ast.TypeAlias, hir.Type]](false, either.Right[*ast.TypeAlias, hir.Type](iter.Value().Second)))
			defer func() {
				self.pkgScope.typeAlias.Set(iter.Value().First, back)
			}()
		}
	}

	for _, f := range node.Fields {
		fn := f.B.Source()
		ft := pair.NewPair(f.A, self.analyseType(f.C))
		st.Fields.Set(fn, ft)
	}
	return st
}

func (self *Analyser) declTrait(node *ast.Trait) {
	methodNameSet := hashset.NewHashSet[string]()
	for _, pair := range node.Methods{
		name := pair.First.Source()
		if methodNameSet.Contain(name){
			errors.ThrowIdentifierDuplicationError(pair.First.Position, pair.First)
		}
		methodNameSet.Add(name)
	}
	if !self.pkgScope.SetTrait(node.Public, node) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
}

func (self *Analyser) defTrait(selfType hir.Type, node *ast.Trait) *hir.TraitDef {
	prevSelfType := self.selfType
	self.selfType = selfType
	defer func() {
		self.selfType = prevSelfType
	}()

	methods := hashmap.NewHashMap[string, *hir.FuncType]()
	for _, pair := range node.Methods{
		methods.Set(pair.First.Source(), self.analyseFuncType(pair.Second))
	}
	return &hir.TraitDef{
		Pkg: stlerror.MustWith(hir.NewPackage(node.Position().Reader.Path().Dir())),
		Name: node.Name.Source(),
		Methods: methods,
	}
}

func (self *Analyser) defTypeAlias(name string) hir.Type {
	res, ok := self.pkgScope.GetTypeAlias("", name)
	if !ok{
		panic("unreachable")
	}
	if t, ok := res.Right(); ok{
		return t
	}

	node, ok := res.Left()
	if !ok{
		panic("unreachable")
	}

	if self.typeAliasTrace.Contain(node){
		errors.ThrowCircularReference(node.Name.Position, node.Name)
	}
	self.typeAliasTrace.Add(node)
	defer func() {
		self.typeAliasTrace.Remove(node)
	}()

	target := self.analyseType(node.Type)
	self.pkgScope.DefTypeAlias(name, target)
	return target
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
	case *ast.StructDef, *ast.Import, *ast.TypeAlias, *ast.Trait, *ast.GenericStructDef:
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
	params := lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
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
	f := &hir.FuncDef{
		Pkg: self.pkgScope.pkg,
		Public:     node.Public,
		ExternName: externName,
		Name:       node.Name.Source(),
		Params:     params,
		Ret:        self.analyseOptionType(node.Ret),
		Body:       util.None[*hir.Block](),
	}
	if f.Name == "main" && !f.Ret.Equal(hir.U8) {
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
	for _, attrObj := range node.Attrs {
		switch attrObj.(type) {
		default:
			panic("unreachable")
		}
	}

	st, ok := self.pkgScope.GetStruct("", node.Scope.Source())
	if !ok{
		errors.ThrowUnknownIdentifierError(node.Scope.Position, node.Scope)
	}

	self.selfType = st
	defer func() {
		self.selfType = nil
	}()

	paramNameSet := hashset.NewHashSetWith[string]()
	params := lo.Map(node.Params, func(paramNode ast.Param, index int) *hir.Param {
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
	f := &hir.MethodDef{
		Pkg: self.pkgScope.pkg,
		Public:    node.Public,
		Scope:     st,
		Name:      node.Name.Source(),
		SelfParam: &hir.Param{
			Mut:  node.ScopeMutable,
			Type: st,
			Name: token.SELFVALUE.String(),
		},
		Params:    params,
		Ret:       self.analyseOptionType(node.Ret),
	}
	if st.Fields.ContainKey(f.Name) || st.Methods.ContainKey(f.Name) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	st.Methods.Set(f.Name, f)
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
	for _, attrObj := range node.Attrs {
		switch attrObj.(type) {
		default:
			panic("unreachable")
		}
	}

	genericParams := linkedhashmap.NewLinkedHashMap[string, *hir.GenericParam]()
	for _, param := range node.GenericParams{
		pn := param.First.Source()
		if genericParams.ContainKey(pn){
			errors.ThrowIdentifierDuplicationError(param.First.Position, param.First)
		}
		genericParam := &hir.GenericParam{Name: pn}

		if constraintNode, ok := param.Second.Value(); ok{
			var pkgName string
			if pkgToken, ok := constraintNode.Pkg.Value(); ok {
				pkgName = pkgToken.Source()
				if !self.pkgScope.externs.ContainKey(pkgName) {
					errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
				}
			}
			traitNode, ok := self.pkgScope.GetTrait(pkgName, constraintNode.Name.Source())
			if !ok{
				errors.ThrowUnknownIdentifierError(constraintNode.Name.Position, constraintNode.Name)
			}
			genericParam.Constraint = util.Some(self.defTrait(genericParam, traitNode))
		}
		genericParams.Set(pn, genericParam)
	}

	f := &hir.GenericFuncDef{
		Pkg: self.pkgScope.pkg,
		Public:     node.Public,
		Name:       node.Name.Source(),
		GenericParams: genericParams,
	}
	scope := _NewFuncScope(self.pkgScope, f)
	self.localScope = scope
	self.genericFuncScope.Set(f, scope)
	defer func() {
		self.localScope = nil
	}()

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

	if !self.pkgScope.SetGenericFunction(f.Name, f) {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
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
	case *ast.StructDef, *ast.Import, *ast.TypeAlias, *ast.Trait, *ast.GenericStructDef:
		return nil
	default:
		panic("unreachable")
	}
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *hir.FuncDef {
	value, ok := self.pkgScope.GetValue("", node.Name.Source())
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
		if !f.Ret.Equal(hir.Empty) {
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
	st, ok := self.pkgScope.GetStruct("", node.Scope.Source())
	if !ok{
		errors.ThrowUnknownIdentifierError(node.Scope.Position, node.Scope)
	}
	f := st.Methods.Get(node.Name.Source())

	self.localScope, self.selfValue, self.selfType = _NewFuncScope(self.pkgScope, f), f.SelfParam, st
	defer func() {
		self.localScope, self.selfValue, self.selfType = nil, nil, nil
	}()

	for i, p := range f.Params {
		if !self.localScope.SetValue(p.Name, p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.Position, node.Params[i].Name)
		}
	}

	var jump hir.BlockEof
	f.Body, jump = self.analyseBlock(node.Body, nil)
	if jump != hir.BlockEofReturn {
		if !f.Ret.Equal(hir.Empty) {
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
	f, ok := self.pkgScope.GetGenericFunction("", node.Name.Source())
	if !ok {
		panic("unreachable")
	}

	self.localScope = self.genericFuncScope.Get(f)
	defer func() {
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
		if !f.Ret.Equal(hir.Empty) {
			errors.ThrowMissingReturnValueError(node.Name.Position, f.Ret)
		}
		f.Body.Stmts.PushBack(&hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		})
	}
	return f
}
