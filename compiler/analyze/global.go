package analyze

import (
	"path/filepath"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"

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
	a.ir.Dependencies = append(a.ir.Dependencies, ir)
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
			a.reporter.Fatalf(
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
	a.ir.Dependencies = append(a.ir.Dependencies, ir)
	return nil
}

func (a *Analyzer) analyzeTypePreDecl(global ast.Global) *globals.TypeDef {
	switch global := global.(type) {
	case *ast.TypeDef:
		decl := globals.NewTypeDef(global.Public, global.Name.OriginText)
		a.typeDef2Ast.Set(decl, global)
		if stlmaps.ContainKey(a.typeName2Def, global.Name.OriginText) {
			a.reporter.Fatalf(
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
	_, ok := a.scope.LookupType(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	decl := a.typeDef2Ast.GetKey(global)
	if !stacks.Add(decl) {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.InvalidRecursionType,
		)
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
	ct, _ := a.scope.LookupType(global.Name.OriginText)
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
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	var t hir.Type
	if tAst, ok := global.Type.Value(); ok {
		t = a.analyzeType(tAst)
	} else {
		v, ok := global.Value.MustValue().(*ast.Func)
		if !ok {
			// TODO: 非函数定义的全局变量的声明解析
			if global.Name.OriginText == "main" {
				a.reporter.Fatalf(
					global.Name.Position,
					report.Errors.InvalidMainFunction,
				)
			}
			return
		} else {
			// 函数定义
			t = a.analyzeFuncDecl(v)
			if global.Name.OriginText == "main" {
				expectType := types.NewFuncType(types.Unit)
				if !t.Equal(expectType) {
					a.reporter.Fatalf(
						global.Name.Position,
						report.Errors.UnexpectedExpression,
						expectType, t,
					)
				}
			}
		}
	}

	let := &locals.Let{
		Pub:      global.Public,
		IsGlobal: true,
		Mut:      global.Mut,
		Type:     t,
		Name:     global.Name.OriginText,
	}

	for _, attrAst := range global.Attributes {
		switch attrAst := attrAst.(type) {
		case *ast.Extern:
			let.ExternalName = optional.Some(attrAst.Name.OriginText)
		default:
			panic("unreachable")
		}
	}

	if bind, ok := global.Bind.Value(); ok {
		bindType := a.analyzeType(bind)
		ct, ok := bindType.(types.CustomType)
		if !ok {
			a.reporter.Fatalf(
				bind.Position(),
				report.Errors.UnexpectedTypeCategory,
				"custom", bindType,
			)
		}
		// TODO: 只能绑定本包定义的类型
		typedef := ct.GetDef()
		_, ok = a.scope.LookupBind(typedef, let.Name)
		if ok {
			a.reporter.Fatalf(
				global.Name.Position,
				report.Errors.RepeatedIdentifier,
				global.Name.OriginText,
			)
		}
		a.scope.AddBind(typedef, let)
	}

	a.scope.AddValue(let)
}

func (a *Analyzer) analyzeGlobalValueDef(global ast.Global) globals.Global {
	switch global := global.(type) {
	case *ast.TypeDef, *ast.Import:
		return nil
	case *ast.Let:
		return a.analyzeLetDef(global, true)
	default:
		panic("unreachable")
	}
}
