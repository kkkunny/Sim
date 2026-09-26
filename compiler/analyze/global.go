package analyze

import (
	"path/filepath"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

// addDependency 追加依赖包（按包指针去重）：重复 import 同一包或显式 import std::buildin
// 时不能产生重复依赖，否则编译 DAG 建边会失败（裸 EdgeDuplicateError）。
func (a *Analyzer) addDependency(pkg *globals.Package) {
	for _, dep := range a.ir.Dependencies {
		if dep == pkg {
			return
		}
	}
	a.ir.Dependencies = append(a.ir.Dependencies, pkg)
}

// 导入buildin包，一定最先导入
func (a *Analyzer) importBuildin() error {
	dirpath := config.BuildinPkgPath
	if a.ir.Path == dirpath {
		return nil
	}

	if _, ok := a.pkgScopes[dirpath]; !ok {
		_, err := analyzeDir(dirpath, a.reporter, a)
		if err != nil {
			return err
		}
	}

	ir, scope := a.pkgScopes[dirpath].Unpack()
	a.scope.Root().AddInclude(scope)
	a.addDependency(ir)
	return nil
}

func (a *Analyzer) analyzeImport(global *ast.Import) error {
	lastPkgToken := stlslices.Last(global.Pkgs)

	name := lastPkgToken.OriginText
	if alias, ok := global.Alias.Value(); ok {
		if alias.Is(token.KindEnum.Dot) {
			name = ""
		} else {
			name = alias.OriginText
		}
	}

	if name != "" {
		_, ok := a.scope.LookupPkg(name)
		if ok {
			a.errorf(
				lastPkgToken.Position,
				report.Errors.RepeatedIdentifier,
				name,
			)
		}
	}

	paths := stlslices.Map(global.Pkgs, func(_ int, tok token.Token) string {
		return tok.OriginText
	})
	dirpath := filepath.Join(append([]string{config.SimRootPath}, paths...)...)

	if _, ok := a.pkgScopes[dirpath]; !ok {
		_, err := analyzeDir(dirpath, a.reporter, a)
		if err != nil {
			return err
		}
	}

	ir, scope := a.pkgScopes[dirpath].Unpack()
	if name != "" {
		a.scope.Root().AddExternal(name, scope)
	} else {
		a.scope.Root().AddInclude(scope)
	}
	a.addDependency(ir)
	return nil
}

func (a *Analyzer) analyzeTypePreDecl(global ast.Global) *globals.TypeDef {
	switch global := global.(type) {
	case *ast.TypeDef:
		decl := globals.NewTypeDef(global.Public, global.Name.OriginText)
		a.typeDef2Ast.Set(decl, global)
		if stlmaps.ContainKey(a.typeName2Def, global.Name.OriginText) {
			a.errorf(
				global.Name.Position,
				report.Errors.RepeatedIdentifier,
				global.Name.OriginText,
			)
		}
		a.typeName2Def[global.Name.OriginText] = decl
		return decl
	default:
		return nil
	}
}

func (a *Analyzer) analyzeTypeDecl(global ast.Global) {
	switch global := global.(type) {
	case *ast.TypeDef:
		a.analyzeCustomTypeDecl(set.StdHashSetWith[*globals.TypeDef](), global)
	}
}

func (a *Analyzer) analyzeCustomTypeDecl(stacks set.Set[*globals.TypeDef], global *ast.TypeDef) types.CustomType {
	// 仅当同名类型来自另一处声明时才算重定义：前向别名（type A B）会在递归中
	// 先注册 B，外层循环再次处理 B 时不应误报
	if exist, ok := a.scope.LookupType(global.Name.OriginText); ok && exist.GetDef() != a.typeDef2Ast.GetKey(global) {
		a.errorf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	decl := a.typeDef2Ast.GetKey(global)
	if decl == nil {
		a.abort()
	}
	if !stacks.Add(decl) {
		a.errorf(
			global.Name.Position,
			report.Errors.CircularReference,
		)
		a.abort()
	}
	defer stacks.Remove(decl)

	var ct types.CustomType
	switch t := global.Type.(type) {
	case *ast.IdentType:
		var underlying hir.Type
		if ct, ok := a.scope.LookupType(t.Name.OriginText); ok {
			underlying = ct
		} else if def, ok := a.typeName2Def[t.Name.OriginText]; ok {
			if a.ir.Path == config.BuildinPkgPath && t.Pkg.IsNone() && global.Name.OriginText == t.Name.OriginText {
				// 允许buildin包内自定义类型名和底层类型同名
				underlying = a.analyzeBuildInIdentType(t)
			} else {
				underlying = a.analyzeCustomTypeDecl(stacks, a.typeDef2Ast.GetValue(def))
			}
		} else {
			underlying = a.analyzeBuildInIdentType(t)
		}
		if _, invalid := underlying.(types.InvalidType); invalid || underlying == nil {
			a.abort()
		}
		switch underlying.(type) {
		case types.SintType:
			ct = types.NewCustomType[types.SintType](decl)
		case types.UintType:
			ct = types.NewCustomType[types.UintType](decl)
		case types.FloatType:
			ct = types.NewCustomType[types.FloatType](decl)
		case types.BooleanType:
			ct = types.NewCustomType[types.BooleanType](decl)
		case types.StringType:
			ct = types.NewCustomType[types.StringType](decl)
		case types.FuncType:
			ct = types.NewCustomType[types.FuncType](decl)
		case types.RefType:
			ct = types.NewCustomType[types.RefType](decl)
		case types.TupleType:
			ct = types.NewCustomType[types.TupleType](decl)
		case types.ArrayType:
			ct = types.NewCustomType[types.ArrayType](decl)
		case types.UnionType:
			ct = types.NewCustomType[types.UnionType](decl)
		case types.StructType:
			ct = types.NewCustomType[types.StructType](decl)
		default:
			panic("unreachable")
		}
	case *ast.FuncType:
		ct = types.NewCustomType[types.FuncType](decl)
	case *ast.RefType:
		ct = types.NewCustomType[types.RefType](decl)
	case *ast.TupleType:
		ct = types.NewCustomType[types.TupleType](decl)
	case *ast.ArrayType:
		ct = types.NewCustomType[types.ArrayType](decl)
	case *ast.UnionType:
		ct = types.NewCustomType[types.UnionType](decl)
	case *ast.StructType:
		ct = types.NewCustomType[types.StructType](decl)
	}

	a.scope.(*scopes.PkgScope).AddType(global.Name.OriginText, ct)
	return ct
}

func (a *Analyzer) analyzeTypeDef(global ast.Global) {
	switch global := global.(type) {
	case *ast.Let, *ast.Import:
		return
	case *ast.TypeDef:
		a.analyzeCustomTypeDef(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeCustomTypeDef(global *ast.TypeDef) *globals.TypeDef {
	ct, ok := a.scope.LookupType(global.Name.OriginText)
	if !ok {
		return nil
	}
	def := ct.GetDef()
	if def.Underlying != nil {
		return def
	}

	if t, ok := global.Type.(*ast.IdentType); ok && a.ir.Path == config.BuildinPkgPath && t.Pkg.IsNone() && global.Name.OriginText == t.Name.OriginText {
		// 允许buildin包内自定义类型名和底层类型同名
		def.Underlying = a.analyzeBuildInIdentType(t)
	} else {
		def.Underlying = a.analyzeType(global.Type)
	}
	return def
}

func (a *Analyzer) analyzeGlobalValueDecl(global ast.Global) {
	switch global := global.(type) {
	case *ast.TypeDef, *ast.Import:
		return
	case *ast.Let:
		a.analyzeGlobalLetDecl(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeGlobalLetDecl(global *ast.Let) {
	_, ok := a.scope.LookupValue(global.Name.OriginText)
	if ok {
		a.errorf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	let := &locals.Let{
		Pub:      global.Public,
		IsGlobal: true,
		Mut:      global.Mut,
		Name:     global.Name.OriginText,
	}

	if bind, ok := global.Bind.Value(); ok {
		bindType := a.analyzeType(bind)
		ct, ok := bindType.(types.CustomType)
		if !ok {
			a.errorf(
				bind.Position(),
				report.Errors.UnexpectedTypeCategory,
				"custom", bindType,
			)
			a.abort()
		}
		// TODO: 只能绑定本包定义的类型
		typedef := ct.GetDef()
		_, ok = a.scope.LookupBind(typedef, let.Name)
		if ok {
			a.errorf(
				global.Name.Position,
				report.Errors.RepeatedIdentifier,
				global.Name.OriginText,
			)
		}
		let.Bind = optional.Some(ct)
		a.scope.AddBind(typedef, let)

		// Self
		selfScope := scopes.NewTemporaryScope(a.scope)
		selfScope.AddType("Self", ct)
		a.scope = selfScope
		defer func() {
			a.scope, _ = a.scope.Parent()
		}()
	}

	if tAst, ok := global.Type.Value(); ok {
		let.Type = a.analyzeType(tAst)
		// 显式标注类型同样要校验 main 签名：否则错误签名会进入 codegen
		//（genEntryWrapper 按 `() -> unit` 调用 sim_main，轻则 ICE、重则静默无入口）
		if global.Name.OriginText == "main" && !isInvalidType(let.Type) {
			a.checkMainType(global, let.Type)
		}
	} else {
		v, ok := global.Value.MustValue().(*ast.Func)
		if ok && !let.Mut {
			// 函数定义
			let.Type = a.analyzeFuncDecl(v)
			if global.Name.OriginText == "main" {
				a.checkMainType(global, let.Type)
			}
		} else {
			if global.Name.OriginText == "main" {
				a.errorf(
					global.Name.Position,
					report.Errors.InvalidMainFunction,
				)
			}
		}
	}

	for _, attrAst := range global.Attributes {
		switch attrAst := attrAst.(type) {
		case *ast.Extern:
			let.ExternalName = optional.Some(attrAst.Name.OriginText)
		default:
			panic("unreachable")
		}
	}

	a.scope.AddValue(let)
	a.letDef2Ast[let] = global
}

// checkMainType 校验入口 main 的签名必须是 `() -> unit`：
// 非函数类型报 InvalidMainFunction，函数但签名不符报 UnexpectedExpression。
func (a *Analyzer) checkMainType(global *ast.Let, t hir.Type) {
	if _, ok := t.(types.FuncType); !ok {
		a.errorf(
			global.Name.Position,
			report.Errors.InvalidMainFunction,
		)
		return
	}
	expectType := types.NewFuncType(types.Unit)
	if !t.Equal(expectType) {
		a.errorf(
			global.Name.Position,
			report.Errors.UnexpectedExpression,
			expectType, t,
		)
	}
}

func (a *Analyzer) analyzeGlobalValueDef(global ast.Global) globals.Global {
	switch global := global.(type) {
	case *ast.TypeDef, *ast.Import:
		return nil
	case *ast.Let:
		a.letDefStack.Clear()
		return a.analyzeGlobalLetDef(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeGlobalLetDef(local *ast.Let) *locals.Let {
	v, ok := a.scope.LookupValue(local.Name.OriginText)
	if !ok {
		// 声明阶段已失败，跳过该全局
		return nil
	}
	decl, ok := v.(*locals.Let)
	if !ok {
		return nil
	}
	if !a.letDefStack.Add(decl) {
		a.errorf(
			local.Name.Position,
			report.Errors.CircularReference,
		)
		a.abort()
	}
	defer a.letDefStack.Remove(decl)

	if decl.Type != nil && (decl.Value.IsSome() || (decl.ExternalName.IsSome() && local.Value.IsNone())) {
		return decl
	}

	// Self
	if bind, ok := decl.Bind.Value(); ok {
		selfScope := scopes.NewTemporaryScope(a.scope)
		selfScope.AddType("Self", bind)
		a.scope = selfScope
		defer func() {
			a.scope, _ = a.scope.Parent()
		}()
	}

	// self
	if !decl.Mut && local.Value.IsSome() && stlval.Is[*ast.Func](local.Value.MustValue()) {
		fast := local.Value.MustValue().(*ast.Func)
		for i, past := range fast.Params {
			if past.Name.OriginText == "self" {
				if i != 0 {
					a.errorf(
						past.Name.Position,
						report.Errors.UnexpectedSelfPosition,
					)
				} else {
					var validSelfType bool
					if ptast, ok := past.Type.(*ast.IdentType); ok && ptast.Pkg.IsNone() && ptast.Name.OriginText == "Self" {
						validSelfType = true
					} else if ptast, ok := past.Type.(*ast.RefType); ok {
						if ptEtAst, ok := ptast.Elem.(*ast.IdentType); ok && ptEtAst.Pkg.IsNone() && ptEtAst.Name.OriginText == "Self" {
							validSelfType = true
						}
					}
					if !validSelfType {
						a.errorf(
							past.Type.Position(),
							report.Errors.UnexpectedSelfType,
						)
					}
				}
			}
		}
	}

	if v, ok := local.Value.Value(); decl.Type != nil && ok {
		decl.Value = optional.Some(a.expectTypeExpr(v, decl.Type))
	} else if decl.Type != nil && decl.ExternalName.IsNone() {
		decl.Value = optional.Some(a.getZeroExpr(local.Name.Position, decl.Type))
	} else {
		v, ok := local.Value.Value()
		if !ok {
			return nil
		}
		decl.Value = optional.Some(a.analyzeExpr(v))
		decl.Type = decl.Value.MustValue().GetType()
	}

	return decl
}
