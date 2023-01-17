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

	Ret    Type
	Params []*Param
	Body   *Block
}

func (self Function) global() {}

func (self Function) stmt() {}

func (self Function) ident() {}

func (self Function) GetType() Type {
	paramTypes := make([]Type, len(self.Params))
	for i, p := range self.Params {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes...)
}

func (self Function) GetMut() bool {
	return false
}

func (self Function) IsTemporary() bool {
	return true
}

func (self Function) IsConst() bool {
	return false
}

func (self Function) GetMethodType() Type {
	paramTypes := make([]Type, len(self.Params)-1)
	for i, p := range self.Params[1:] {
		paramTypes[i] = p.Type
	}
	return NewFuncType(self.Ret, paramTypes...)
}

// GlobalVariable 全局变量
type GlobalVariable struct {
	ExternName string

	Type  Type
	Value Expr
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

func (self GlobalVariable) IsConst() bool {
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

// 全局变量
func analyseGlobalVariable(ctx *packageContext, ast *parse.GlobalValue) (*GlobalVariable, utils.Error) {
	if ast.Variable.Type == nil && ast.Variable.Value == nil {
		return nil, utils.Errorf(ast.Variable.Name.Pos, "expect a type or a value")
	}

	var typ Type
	var err utils.Error
	if ast.Variable.Type != nil {
		typ, err = analyseType(ctx, ast.Variable.Type)
		if err != nil {
			return nil, err
		}
	}

	var value Expr
	if ast.Variable.Type != nil && ast.Variable.Value != nil {
		value, err = expectExpr(newBlockContext(newFunctionContext(ctx, None), false), typ, ast.Variable.Value)
		if err != nil {
			return nil, err
		}
		if !value.IsConst() {
			return nil, utils.Errorf(ast.Variable.Value.Position(), "expect a constant value")
		}
	} else if ast.Variable.Type == nil && ast.Variable.Value != nil {
		value, err = analyseExpr(newBlockContext(newFunctionContext(ctx, None), false), nil, ast.Variable.Value)
		if err != nil {
			return nil, err
		} else if IsNoneType(value.GetType()) {
			return nil, utils.Errorf(ast.Variable.Value.Position(), "expect a value")
		} else if !value.IsConst() {
			return nil, utils.Errorf(ast.Variable.Value.Position(), "expect a constant value")
		}
		typ = value.GetType()
	}

	v := &GlobalVariable{
		Type:  typ,
		Value: value,
	}
	if !ctx.AddValue(ast.Public, ast.Variable.Name.Source, v) {
		return nil, utils.Errorf(ast.Variable.Name.Pos, "duplicate identifier")
	}

	// 属性
	var errors []utils.Error
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
					errors = append(errors, utils.Errorf(asm.Position(), "can not find path `%s`", linkPath))
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
	if v.ExternName == "" && ast.Variable.Value == nil {
		errors = append(errors, utils.Errorf(ast.Variable.Name.Pos, "missing value"))
	}
	if len(errors) == 0 {
		return v, nil
	} else if len(errors) == 1 {
		return nil, errors[0]
	} else {
		return nil, utils.NewMultiError(errors...)
	}
}

// 方法声明
func analyseMethodDecl(ctx *packageContext, ast *parse.Method) (*Function, utils.Error) {
	_selfType, err := analyseType(ctx, parse.NewTypeIdent(nil, ast.Self))
	if err != nil {
		return nil, err
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
	return f, nil
}

// 方法定义
func analyseMethodDef(ctx *packageContext, ast *parse.Method) utils.Error {
	_selfType, err := analyseType(ctx, parse.NewTypeIdent(nil, ast.Self))
	if err != nil {
		return err
	}

	name := _selfType.String() + "." + ast.Name.Source
	f := ctx.GetValue(name).Second.(*Function)
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
