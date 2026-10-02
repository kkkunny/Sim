package llgen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (lg *LLvmGenerator) genTypeDecl(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		lg.genCustomTypeDecl(global)
	}
}

func (lg *LLvmGenerator) genTypeDef(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		lg.genCustomTypeDef(global)
	}
}

// 预先注册全局符号名，允许全局变量前向引用
func (lg *LLvmGenerator) genGlobalDecl(global globals.Global) {
	switch global := global.(type) {
	case *locals.Let:
		lg.genGlobalLetDecl(global)
	}
}

func (lg *LLvmGenerator) genGlobalLetDecl(l *locals.Let) {
	var name string
	if externalName, ok := l.ExternalName.Value(); ok {
		name = externalName
	} else if l.Name == "main" {
		name = "sim_main"
	} else {
		name = stableName(lg.pkg, l.Name)
	}
	lg.ctx.idents[l] = &Ident{Name: name}
}

func (lg *LLvmGenerator) genGlobalValue(global globals.Global) {
	switch global := global.(type) {
	case *locals.Let:
		lg.genGlobalLet(global)
	}
}

func (lg *LLvmGenerator) genGlobalLet(l *locals.Let) {
	ident := lg.ctx.idents[l]
	if isFuncLet(l) {
		lg.genGlobalFuncLet(l, ident)
		return
	}

	t := lg.genType(l.GetType())
	g := lg.getOrDeclareVar(l, ident.Name, t)
	if v, ok := l.Value.Value(); ok {
		if c, ok := lg.genConstExpr(v); ok {
			g.SetInitializer(c)
		} else {
			g.SetInitializer(lg.ctx.llvmCtx.ConstZero(t.DynType()))
			lg.genGlobalVarCtor(g, v)
		}
	} else {
		g.SetInitializer(lg.ctx.llvmCtx.ConstZero(t.DynType()))
	}
}

func (lg *LLvmGenerator) genGlobalFuncLet(l *locals.Let, ident *Ident) {
	ft := l.GetType().(types.FuncType)
	if externalName, ok := l.ExternalName.Value(); ok && l.Value.IsNone() {
		lg.getOrDeclareFunc(externalName, ft)
		return
	}

	value, hasValue := l.Value.Value()
	if !hasValue {
		return
	}
	expr, ok := value.(*locals.Func)
	if !ok {
		return
	}
	fn := lg.getOrDeclareFunc(ident.Name, ft)
	if l.Name == "main" || (l.ExternalName.IsNone() && !l.Pub) {
		fn.SetLinkage(llvm.LinkageInternal)
	}
	if b, ok := expr.Body.Value(); ok {
		lg.genFunctionBody(fn, expr, func() {
			lg.genNativeFuncParams(fn, expr.Params, 0)
			lg.genBlockStmts(b)
			lg.ensureTerminator(ft.GetReturn())
		})
	}
}

// 是否是函数定义/声明
func isFuncLet(l *locals.Let) bool {
	if l.Mut || !stlval.Is[types.FuncType](l.GetType()) {
		return false
	}
	if v, ok := l.Value.Value(); ok {
		return stlval.Is[*locals.Func](v)
	}
	return l.ExternalName.IsSome()
}

func (lg *LLvmGenerator) getOrDeclareFunc(name string, ft types.FuncType) ir.Function {
	fn, _ := lg.model.GetOrCreateFunction(name, lg.genNativeFuncType(ft))
	return fn
}

func (lg *LLvmGenerator) getOrDeclareVar(l *locals.Let, name string, t llvm.AnyType) ir.Global {
	g, created := lg.model.GetOrCreateGlobal(name, t)
	if created && !l.Pub && l.ExternalName.IsNone() {
		g.SetLinkage(llvm.LinkageInternal)
	}
	return g
}

// 生成main函数包装
func (lg *LLvmGenerator) genMainFunc() {
	simMain, ok := lg.model.GetFunction("sim_main")
	if !ok {
		return
	}
	llctx := lg.ctx.llvmCtx
	mainFn := lg.model.NewFunction("main", llctx.Fn(llctx.Int(32), nil, false))
	lg.genFunctionBody(mainFn, nil, func() {
		lg.builder.Call[llvm.DynT](simMain, nil, "")
		lg.builder.Ret(llctx.Int(32).Const(0))
	})
}

// genConstExpr 生成全局变量初始化常量；非常量返回false
func (lg *LLvmGenerator) genConstExpr(expr locals.Expr) (llvm.AnyValue, bool) {
	llctx := lg.ctx.llvmCtx
	switch expr := expr.(type) {
	case *locals.Integer:
		return lg.genInteger(expr.Value, expr.GetType()), true
	case *locals.Float:
		v, _ := expr.Value.Float64()
		return llvm.MustFloatType(lg.genType(expr.GetType())).Const(v), true
	case *locals.Boolean:
		return llctx.ConstBool(expr.Value), true
	case *locals.String:
		return lg.genString(expr.Value), true
	case *locals.Func:
		if len(expr.CaptureVariables) > 0 {
			return nil, false
		}
		v, _, _ := lg.genFuncValue(expr)
		return v, v.IsConstant()
	case *locals.Tuple:
		elems, ok := lg.genConstElems(expr.Elems)
		if !ok {
			return nil, false
		}
		return lg.buildAggregateConst(expr.GetType(), elems), true
	case *locals.Array:
		elems, ok := lg.genConstElems(expr.Elems)
		if !ok {
			return nil, false
		}
		elemT := lg.genType(expr.Type.GetElem())
		return lg.buildAggregateConst(expr.GetType(), []llvm.AnyValue{llctx.ConstArray(elemT, elems...)}), true
	case *locals.Struct:
		elems := make([]llvm.AnyValue, 0, len(expr.Type.GetFields()))
		for _, field := range expr.Type.GetFields() {
			fieldValue, ok := expr.Fields[field.Name]
			if !ok {
				elems = append(elems, llctx.ConstZero(lg.genType(field.Type).DynType()))
				continue
			}
			v, ok := lg.genConstExpr(fieldValue)
			if !ok {
				return nil, false
			}
			elems = append(elems, v)
		}
		return lg.buildAggregateConst(expr.GetType(), elems), true
	default:
		return nil, false
	}
}

// genConstElems 生成常量元素列表；任一元素非常量时返回false
func (lg *LLvmGenerator) genConstElems(exprs []locals.Expr) ([]llvm.AnyValue, bool) {
	elems := make([]llvm.AnyValue, len(exprs))
	for i, e := range exprs {
		v, ok := lg.genConstExpr(e)
		if !ok {
			return nil, false
		}
		elems[i] = v
	}
	return elems, true
}

// buildAggregateConst 构造聚合类型常量（自定义类型使用命名结构体）
func (lg *LLvmGenerator) buildAggregateConst(t hir.Type, elems []llvm.AnyValue) llvm.AnyValue {
	structType := llvm.MustStructType(lg.genType(t))
	if structType.Name() == "" {
		return lg.ctx.llvmCtx.ConstStruct(false, elems...)
	}
	return lg.ctx.llvmCtx.ConstNamedStruct(structType, elems...)
}

// genGlobalVarCtor 生成全局变量动态初始化构造器
func (lg *LLvmGenerator) genGlobalVarCtor(g ir.Global, value locals.Expr) {
	llctx := lg.ctx.llvmCtx
	fn := lg.model.NewFunction(lg.uniqueFuncName(), llctx.Fn(llctx.Void(), nil, false))
	fn.SetLinkage(llvm.LinkageInternal)
	lg.genFunctionBody(fn, nil, func() {
		v := lg.genExpr(value)
		lg.store(v, g.Value)
		lg.builder.RetVoid()
	})
	lg.model.AppendCtor(fn, 65535)
}
