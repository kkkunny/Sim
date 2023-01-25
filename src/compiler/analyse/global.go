package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
)

// Global 全局
type Global interface {
	global()
}

// Function 函数
type Function struct {
	// 属性
	ExternName string // 外部名
	NoReturn   bool   // 函数是否不返回
	Inline     *bool  // 函数是否强制内联或者强制不内联
	Init, Fini bool   // 是否是init or fini函数

	Ret    Type
	Params []*Param
	VarArg bool
	Body   *Block // 可能为空
}

func (self Function) global() {}

func (self Function) stmt() {}

func (self Function) ident() {}

func (self Function) GetType() Type {
	paramTypes := make([]Type, len(self.Params))
	for i, p := range self.Params {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes, self.VarArg)
}

func (self Function) GetMut() bool {
	return false
}

func (self Function) IsTemporary() bool {
	return true
}

func (self Function) GetMethodType() *TypeFunc {
	paramTypes := make([]Type, len(self.Params)-1)
	for i, p := range self.Params[1:] {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes, self.VarArg)
}

// GlobalVariable 全局变量
type GlobalVariable struct {
	ExternName string

	Type  Type
	Value Expr // 可能为空
}

func (self GlobalVariable) global() {}

func (self GlobalVariable) stmt() {}

func (self GlobalVariable) ident() {}

func (self GlobalVariable) GetType() Type {
	return self.Type
}

func (self GlobalVariable) GetMut() bool {
	return true
}

func (self GlobalVariable) IsTemporary() bool {
	return false
}

// *********************************************************************************************************************

// 外部函数声明
func analyseExternFunction(ctx *packageContext, ast *parse.ExternFunction) (*Function, utils.Error) {
	retType, err := analyseType(ctx, ast.Ret)
	if err != nil {
		return nil, err
	}

	params := make([]*Param, len(ast.Params))
	var errors []utils.Error
	for i, p := range ast.Params {
		pt, err := analyseType(ctx, p.Type)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		params[i] = &Param{Type: pt}
	}
	if len(errors) == 1 {
		return nil, errors[0]
	} else if len(errors) > 1 {
		return nil, utils.NewMultiError(errors...)
	}

	f := &Function{
		Ret:    retType,
		Params: params,
		VarArg: ast.VarArg,
	}

	// 属性
	errors = make([]utils.Error, 0)
	for _, astAttr := range ast.Attrs {
		switch attr := astAttr.(type) {
		case *parse.AttrExtern:
			f.ExternName = attr.Name.Source
		case *parse.AttrLink:
			for _, asm := range attr.Asms {
				linkPath := stlos.Path(asm.Value)
				if !linkPath.IsAbsolute() {
					linkPath = ctx.path.Join(linkPath)
				}
				if !linkPath.IsExist() {
					errors = append(errors, utils.Errorf(asm.Position(), "can not find path `%s`", linkPath))
				}
				ctx.f.Links[linkPath] = struct{}{}
			}
			for _, lib := range attr.Libs {
				ctx.f.Libs[lib.Value] = struct{}{}
			}
		case *parse.AttrNoReturn:
			f.NoReturn = true
		case *parse.AttrInit:
			f.Init = true
		case *parse.AttrFini:
			f.Fini = true
		default:
			panic("unknown attr")
		}
	}
	if len(errors) == 1 {
		return nil, errors[0]
	} else if len(errors) > 1 {
		return nil, utils.NewMultiError(errors...)
	}

	if !ctx.AddValue(ast.Public, ast.Name.Source, f) {
		return nil, utils.Errorf(ast.Name.Pos, "duplicate identifier")
	}
	return f, nil
}

// 函数声明
func analyseFunctionDecl(ctx *packageContext, ast *parse.Function) (*Function, utils.Error) {
	retType, err := analyseType(ctx, ast.Ret)
	if err != nil {
		return nil, err
	}

	params := make([]*Param, len(ast.Params))
	var errors []utils.Error
	for i, p := range ast.Params {
		pt, err := analyseType(ctx, p.Type)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		params[i] = &Param{Type: pt}
	}
	if len(errors) == 1 {
		return nil, errors[0]
	} else if len(errors) > 1 {
		return nil, utils.NewMultiError(errors...)
	}

	f := &Function{
		Ret:    retType,
		Params: params,
		VarArg: ast.VarArg,
	}

	// 属性
	for _, astAttr := range ast.Attrs {
		switch attr := astAttr.(type) {
		case *parse.AttrExtern:
			f.ExternName = attr.Name.Source
		case *parse.AttrNoReturn:
			f.NoReturn = true
		case *parse.AttrInline:
			var v bool
			if attr.Value.Kind == lex.TRUE {
				v = true
			} else {
				v = false
			}
			f.Inline = &v
		case *parse.AttrInit:
			f.Init = true
		case *parse.AttrFini:
			f.Fini = true
		default:
			panic("unknown attr")
		}
	}

	if !ctx.AddValue(ast.Public, ast.Name.Source, f) {
		return nil, utils.Errorf(ast.Name.Pos, "duplicate identifier")
	}
	return f, nil
}

// 函数定义
func analyseFunctionDef(ctx *packageContext, f *Function, ast *parse.Function) utils.Error {
	fctx := newFunctionContext(ctx, f.Ret)
	for i, p := range f.Params {
		name := ast.Params[i].Name
		if name != nil {
			if !fctx.AddValue(name.Source, p) {
				return utils.Errorf(name.Pos, "duplicate identifier")
			}
		}
	}

	bctx, body, err := analyseBlock(fctx, ast.Body, false)
	if err != nil {
		return err
	} else if !bctx.IsEnd() {
		if f.Ret.Equal(None) {
			body.Stmts = append(body.Stmts, &Return{})
			bctx.SetEnd()
		} else {
			return utils.Errorf(ast.Name.Pos, "function missing return")
		}
	}
	f.Body = body
	return nil
}

// 全局变量声明
func analyseGlobalVariableDecl(ctx *packageContext, ast *parse.GlobalValue) (*GlobalVariable, utils.Error) {
	typ, err := analyseType(ctx, ast.Type)
	if err != nil {
		return nil, err
	}

	v := &GlobalVariable{
		Type: typ,
	}

	// 属性
	var errs []utils.Error
	for _, astAttr := range ast.Attrs {
		switch attr := astAttr.(type) {
		case *parse.AttrExtern:
			v.ExternName = attr.Name.Source
		case *parse.AttrLink:
			for _, asm := range attr.Asms {
				linkPath := stlos.Path(asm.Value)
				if !linkPath.IsAbsolute() {
					linkPath = ctx.path.Join(linkPath)
				}
				if !linkPath.IsExist() {
					errs = append(errs, utils.Errorf(asm.Position(), "can not find path `%s`", linkPath))
				}
				ctx.f.Links[linkPath] = struct{}{}
			}
			for _, lib := range attr.Libs {
				ctx.f.Libs[lib.Value] = struct{}{}
			}
		default:
			panic("unknown attr")
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	if !ctx.AddValue(ast.Public, ast.Name.Source, v) {
		return nil, utils.Errorf(ast.Name.Pos, "duplicate identifier")
	}
	return v, nil
}

// 全局变量定义
func analyseGlobalVariableDef(ctx *packageContext, v *GlobalVariable, ast *parse.GlobalValue) utils.Error {
	var err utils.Error
	if ast.Value != nil {
		v.Value, err = expectExpr(newBlockContext(newFunctionContext(ctx, None), false), v.Type, ast.Value)
	} else if v.ExternName == "" {
		v.Value, err = getDefaultExprByType(ast.Type.Position(), v.Type)
	}
	return err
}

// 方法声明
func analyseMethodDecl(ctx *packageContext, ast *parse.Method) (*Function, utils.Error) {
	_selfTypeObj, err := analyseType(ctx, parse.NewTypeIdent(nil, ast.Self))
	if err != nil {
		return nil, err
	}
	_selfType, ok := _selfTypeObj.(*Typedef)
	if !ok {
		return nil, utils.Errorf(ast.Self.Pos, "expect a typedef")
	}
	selfType := NewPtrType(_selfType)

	retType, err := analyseType(ctx, ast.Ret)
	if err != nil {
		return nil, err
	}

	params := make([]*Param, len(ast.Params)+1)
	params[0] = &Param{Type: selfType}
	var errors []utils.Error
	for i, p := range ast.Params {
		pt, err := analyseType(ctx, p.Type)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		params[i+1] = &Param{Type: pt}
	}
	if len(errors) == 1 {
		return nil, errors[0]
	} else if len(errors) > 1 {
		return nil, utils.NewMultiError(errors...)
	}

	f := &Function{
		Ret:    retType,
		Params: params,
		VarArg: ast.VarArg,
	}

	// 属性
	for _, astAttr := range ast.Attrs {
		switch attr := astAttr.(type) {
		case *parse.AttrNoReturn:
			f.NoReturn = true
		case *parse.AttrInline:
			var v bool
			if attr.Value.Kind == lex.TRUE {
				v = true
			} else {
				v = false
			}
			f.Inline = &v
		default:
			panic("unknown attr")
		}
	}

	name := _selfType.String() + "." + ast.Name.Source
	if !ctx.AddValue(ast.Public, name, f) {
		return nil, utils.Errorf(ast.Name.Pos, "duplicate identifier")
	}
	_selfType.Methods[ast.Name.Source] = f
	return f, nil
}

// 方法定义
func analyseMethodDef(ctx *packageContext, ast *parse.Method) utils.Error {
	_selfTypeObj, err := analyseType(ctx, parse.NewTypeIdent(nil, ast.Self))
	if err != nil {
		return err
	}
	_selfType := _selfTypeObj.(*Typedef)

	f := _selfType.Methods[ast.Name.Source]
	fctx := newFunctionContext(ctx, f.Ret)
	for i, p := range f.Params {
		if i == 0 {
			fctx.AddValue("self", p)
		} else {
			pn := ast.Params[i-1].Name
			if pn != nil {
				if !fctx.AddValue(pn.Source, p) {
					return utils.Errorf(pn.Pos, "duplicate identifier")
				}
			}
		}
	}

	bctx, body, err := analyseBlock(fctx, ast.Body, false)
	if err != nil {
		return err
	} else if !bctx.IsEnd() {
		if f.Ret.Equal(None) {
			body.Stmts = append(body.Stmts, &Return{})
			bctx.SetEnd()
		} else {
			return utils.Errorf(ast.Name.Pos, "function missing return")
		}
	}
	f.Body = body
	return nil
}
