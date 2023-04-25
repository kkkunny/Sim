package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
)

// 类型定义
func (self *Analyser) analyseTypedef(ast parse.TypeDef) (*hir.Typedef, utils.Error) {
	def, ok := self.symbol.lookupType(ast.Name.Source)
	if !ok {
		panic("unreachable")
	}

	elem, err := self.analyseType(ast.Target)
	if err != nil {
		return nil, err
	}

	self.symbol.defType(ast.Name.Source, elem)

	return def.data, nil
}

// 全局变量声明
func (self *Analyser) analyseGlobalValueDecl(ast parse.GlobalValue) (*hir.GlobalValue, utils.Error) {
	// 类型
	typ, err := self.analyseType(ast.Type)
	if err != nil {
		return nil, err
	}

	// 声明
	ident := hir.NewGlobalValue(ast.Mutable, typ, "", nil)
	if !self.symbol.defValue(ast.Public, ast.Name.Source, ident) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	self.globals.PushBack(ident)

	// 属性
	for _, attrObj := range ast.Attrs {
		switch attr := attrObj.(type) {
		case *parse.AttrExtern:
			// TODO: 链接名重复
			ident.Name = attr.Name.Source
		default:
			panic("unreachable")
		}
	}

	return ident, nil
}

// 函数声明
func (self *Analyser) analyseFunctionDecl(ast parse.Function) (*hir.Function, utils.Error) {
	// 类型
	ret, err := self.analyseType(ast.Ret)
	if err != nil {
		return nil, err
	}
	params := make([]hir.Type, len(ast.Params))
	var errs []utils.Error
	for i, p := range ast.Params {
		params[i], err = self.analyseType(p.Type)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}
	ft := hir.NewTypeFunc(ast.VarArg, ret, params...)

	// 声明
	ident := hir.NewFunction(ft, "", nil, nil)
	if !self.symbol.defValue(ast.Public, ast.Name.Source, ident) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	self.globals.PushBack(ident)

	// 属性
	for _, attrObj := range ast.Attrs {
		switch attr := attrObj.(type) {
		case *parse.AttrExtern:
			// TODO: 链接名重复
			ident.Name = attr.Name.Source
		case *parse.AttrNoReturn:
			ident.NoReturn = true
		case *parse.AttrInline:
			if attr.Value.Kind == lex.TRUE {
				ident.MustInline = true
			} else {
				ident.MustNoInline = true
			}
		case *parse.AttrInit:
			ident.Init = true
		case *parse.AttrFini:
			ident.Fini = true
		default:
			panic("unreachable")
		}
	}

	return ident, nil
}

// 泛型函数声明
func (self *Analyser) analyseGenericFunctionDecl(ast parse.Function) utils.Error {
	// 泛型参数
	var errs []utils.Error
	nameSet := make(map[string]struct{})
	for _, p := range ast.GenericParams {
		if _, ok := nameSet[p.Source]; ok {
			errs = append(errs, utils.Errorf(p.Pos, errDuplicateDeclaration))
		} else {
			nameSet[p.Source] = struct{}{}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 声明
	if !self.symbol.defGenericFunc(ast.Public, ast.Name.Source, ast) {
		return utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	return nil
}

// 方法声明
func (self *Analyser) analyseMethodDecl(ast parse.Method) (*hir.Method, utils.Error) {
	// self
	_t, ok := self.symbol.lookupType(ast.Self.Source)
	if !ok {
		return nil, utils.Errorf(ast.Self.Pos, errUnknownIdentifier)
	}
	selfDef := _t.data

	// 类型
	ret, err := self.analyseType(ast.Ret)
	if err != nil {
		return nil, err
	}
	params := make([]hir.Type, len(ast.Params))
	var errs []utils.Error
	for i, p := range ast.Params {
		params[i], err = self.analyseType(p.Type)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}
	ft := hir.NewTypeFunc(ast.VarArg, ret, params...)

	// 声明
	ident := hir.NewMethod(ft, ast.Mutable, selfDef, nil, nil)
	if selfDef.Target.Kind == hir.TStruct {
		for _, f := range selfDef.Target.GetStructFields() {
			if f.Second == ast.Name.Source {
				return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
			}
		}
	}
	if !selfDef.DeclMethod(ast.Name.Source, ident) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	self.globals.PushBack(ident)

	// 属性
	for _, attrObj := range ast.Attrs {
		switch attr := attrObj.(type) {
		case *parse.AttrNoReturn:
			ident.NoReturn = true
		case *parse.AttrInline:
			if attr.Value.Kind == lex.TRUE {
				ident.MustInline = true
			} else {
				ident.MustNoInline = true
			}
		default:
			panic("unreachable")
		}
	}

	return ident, nil
}

// 泛型方法声明
func (self *Analyser) analyseGenericMethodDecl(ast parse.Method) utils.Error {
	if _t, ok := self.symbol.lookupType(ast.Self.Source); ok {
		return self.analyseGenericMethodDeclForTypedef(_t.data, ast)
	} else if _t, ok := self.symbol.lookupGenericType(ast.Self.Source); ok {
		return self.analyseGenericMethodDeclForGenericTypedef(_t.data, ast)
	} else {
		return utils.Errorf(ast.Self.Pos, errUnknownIdentifier)
	}
}

// 类型定义的泛型方法声明
func (self *Analyser) analyseGenericMethodDeclForTypedef(selfDef *hir.Typedef, ast parse.Method) utils.Error {
	// 泛型参数
	var errs []utils.Error
	nameSet := make(map[string]struct{})
	for _, p := range ast.GenericParams {
		if _, ok := nameSet[p.Source]; ok {
			errs = append(errs, utils.Errorf(p.Pos, errDuplicateDeclaration))
		} else {
			nameSet[p.Source] = struct{}{}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 声明
	if !self.symbol.defGenericMethod(ast.Public, selfDef.Name, ast.Name.Source, ast) {
		return utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	return nil
}

// 泛型类型定义的泛型方法声明
func (self *Analyser) analyseGenericMethodDeclForGenericTypedef(defAst parse.TypeDef, ast parse.Method) utils.Error {
	// 泛型参数
	if len(ast.SelfGenericParams) != len(defAst.GenericParams) {
		return utils.Errorf(
			ast.Self.Pos,
			"expect `%d` generic params but there is `%s`",
			len(defAst.GenericParams),
			len(ast.SelfGenericParams),
		)
	}
	var errs []utils.Error
	nameSet := make(map[string]struct{})
	for i, sp := range ast.SelfGenericParams {
		if defAst.GenericParams[i].Source != sp.Source {
			errs = append(
				errs,
				utils.Errorf(
					sp.Pos,
					"expect name `%s` but there is `%s`",
					defAst.GenericParams[i].Source,
					sp.Source,
				),
			)
		} else {
			nameSet[sp.Source] = struct{}{}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}
	for _, p := range ast.GenericParams {
		if _, ok := nameSet[p.Source]; ok {
			errs = append(errs, utils.Errorf(p.Pos, errDuplicateDeclaration))
		} else {
			nameSet[p.Source] = struct{}{}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 声明
	if !self.symbol.defGenericMethod(ast.Public, defAst.Name.Source, ast.Name.Source, ast) {
		return utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	return nil
}

// 全局变量定义
func (self *Analyser) analyseGlobalValueDef(ast parse.GlobalValue) utils.Error {
	_ident, ok := self.symbol.lookupValue(ast.Name.Source)
	if !ok {
		panic("unreachable")
	}
	ident := _ident.data.(*hir.GlobalValue)

	if ident.Name != "" && ast.Value == nil {
		return nil
	}

	// 值
	var value hir.Expr
	var err utils.Error
	if ast.Value == nil {
		value, err = self.getDefaultValue(ident.Typ)
	} else {
		value, err = self.expectExpr(ident.Typ, ast.Value)
	}
	if err != nil {
		return err
	}

	ident.Value = value
	return nil
}

// 函数定义
func (self *Analyser) analyseFunctionDef(ast parse.Function) utils.Error {
	if ast.Body == nil {
		return nil
	}

	_ident, ok := self.symbol.lookupValue(ast.Name.Source)
	if !ok {
		panic("unreachable")
	}
	ident := _ident.data.(*hir.Function)
	ft := ident.Type()

	// 参数
	ident.Params = make([]*hir.Param, len(ast.Params))
	pro := func(symbol *symbolTable) utils.Error {
		paramTypes := ft.GetFuncParams()
		for i, p := range ast.Params {
			ident.Params[i] = hir.NewParam(p.Mutable, paramTypes[i])
			if p.Name != nil {
				symbol.defValue(false, p.Name.Source, ident.Params[i])
			}
		}
		symbol.ret = ft.GetFuncRet()
		return nil
	}

	// 函数体
	block, ret, err := self.analyseBlock(*ast.Body, pro)
	if err != nil {
		return err
	} else if !ret {
		if !ft.GetFuncRet().IsNone() {
			return utils.Errorf(ast.Name.Pos, "expect a return value")
		}
		block.Stmts.PushBack(hir.NewReturn(nil))
	}
	ident.Body = block

	return nil
}

// 方法定义
func (self *Analyser) analyseMethodDef(ast parse.Method) utils.Error {
	// self
	_t, ok := self.symbol.lookupType(ast.Self.Source)
	if !ok {
		panic("unreachable")
	}
	selfDef := _t.data

	ident, ok := selfDef.LookupMethod(ast.Name.Source)
	if !ok {
		panic("unreachable")
	}
	ft := ident.FunctionType()

	// 参数
	ident.Params = make([]*hir.Param, len(ast.Params)+1)
	pro := func(symbol *symbolTable) utils.Error {
		paramTypes := ft.GetFuncParams()
		ident.Params[0] = hir.NewParam(ast.Mutable, paramTypes[0])
		symbol.defValue(false, "self", ident.Params[0])
		paramTypes = paramTypes[1:]
		for i, p := range ast.Params {
			ident.Params[i+1] = hir.NewParam(p.Mutable, paramTypes[i])
			if p.Name != nil {
				symbol.defValue(false, p.Name.Source, ident.Params[i+1])
			}
		}
		symbol.ret = ft.GetFuncRet()
		return nil
	}

	// 函数体
	block, ret, err := self.analyseBlock(*ast.Body, pro)
	if err != nil {
		return err
	} else if !ret {
		if !ft.GetFuncRet().IsNone() {
			return utils.Errorf(ast.Name.Pos, "expect a return value")
		}
		block.Stmts.PushBack(hir.NewReturn(nil))
	}
	ident.Body = block

	return nil
}

// 泛型类型声明
func (self *Analyser) analyseGenericTypeDecl(ast parse.TypeDef) utils.Error {
	// 泛型参数
	var errs []utils.Error
	nameSet := make(map[string]struct{})
	for _, p := range ast.GenericParams {
		if _, ok := nameSet[p.Source]; ok {
			errs = append(errs, utils.Errorf(p.Pos, errDuplicateDeclaration))
		} else {
			nameSet[p.Source] = struct{}{}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 声明
	if !self.symbol.defGenericType(ast.Public, ast.Name.Source, ast) {
		return utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	return nil
}
