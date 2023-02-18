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
	delete(self.typedefTemp, ast.Name.Source)

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
	ident := hir.NewGlobalValue(typ, "", nil)
	if !self.symbol.defValue(ast.Public, ast.Name.Source, ident) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}

	// 属性
	for _, attrObj := range ast.Attrs {
		switch attr := attrObj.(type) {
		case *parse.AttrExtern:
			// TODO: 链接名重复
			ident.Name = attr.Name.Source
		case *parse.AttrLink:
			// TODO: 链接
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

	// 属性
	for _, attrObj := range ast.Attrs {
		switch attr := attrObj.(type) {
		case *parse.AttrExtern:
			// TODO: 链接名重复
			ident.Name = attr.Name.Source
		case *parse.AttrLink:
			// TODO: 链接
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
	ident := hir.NewMethod(ft, selfDef, nil, nil)
	if !selfDef.DeclMethod(ast.Name.Source, ident) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}

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

// 全局变量定义
func (self *Analyser) analyseGlobalValueDef(ast parse.GlobalValue) utils.Error {
	_ident, ok := self.symbol.lookupValue(ast.Name.Source)
	if !ok {
		panic("unreachable")
	}
	ident := _ident.data.(*hir.GlobalValue)

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
			if p.Name != nil {
				ident.Params[i] = hir.NewParam(paramTypes[i])
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
		ident.Params[0] = hir.NewParam(paramTypes[0])
		symbol.defValue(false, "self", ident.Params[0])
		paramTypes = paramTypes[1:]
		for i, p := range ast.Params {
			if p.Name != nil {
				ident.Params[i+1] = hir.NewParam(paramTypes[i])
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
