package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/reader"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func (self *Analyser) analyseImport(node *ast.Import) *global.Package {
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
	pkgPath := config.OfficialPkgPath.Join(paths...)

	// 导入包
	pkg, err := self.importPackage(pkgPath, pkgName, importAll)
	if err != nil {
		switch e := err.(type) {
		case *importPackageCircularError:
			errors.ThrowPackageCircularReference(stlslices.Last(node.Paths).Position, stlslices.Map(e.chain, func(i int, pkg *global.Package) stlos.FilePath {
				return pkg.Path()
			}))
		case *importPackageDuplicationError:
			errors.ThrowIdentifierDuplicationError(stlslices.Last(node.Paths).Position, stlslices.Last(node.Paths))
		case *importPackageInvalidError:
			errors.ThrowInvalidPackage(reader.MixPosition(stlslices.First(node.Paths).Position, stlslices.Last(node.Paths).Position), node.Paths)
		default:
			panic("unreachable")
		}
	}
	return pkg
}

func (self *Analyser) declTrait(node *ast.Trait) {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name, true)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	self.pkg.AppendGlobal(node.Public, global.NewTrait(name, hashmap.StdWithCap[string, *global.FuncDecl](uint(len(node.Methods)))))
}

func (self *Analyser) declTypeDef(node *ast.TypeDef) {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name, true)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	self.pkg.AppendGlobal(node.Public, global.NewCustomTypeDef(name, types.NoThing))
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name, true)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	self.pkg.AppendGlobal(node.Public, global.NewAliasTypeDef(name, types.NoThing))
}

func (self *Analyser) defTypeDef(node *ast.TypeDef, typedefs hashmap.HashMap[string, types.Type]) {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(*global.CustomTypeDef)
	// TODO: 无法确定Self、TypeDef类型
	decl.SetTarget(self.analyseType(node.Target, self.structTypeAnalyser(), self.enumTypeAnalyser()))
}

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(*global.AliasTypeDef)
	// TODO: 无法确定TypeDef类型
	decl.SetTarget(self.analyseType(node.Target))
}

func (self *Analyser) defTrait(node *ast.Trait) {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(*global.Trait)

	for _, methodNode := range node.Methods {
		method := self.analyseFuncDecl(*methodNode, self.selfTypeAnalyser(true))
		if decl.Methods.Contain(method.Name()) {
			errors.ThrowIdentifierDuplicationError(methodNode.Name.Position, methodNode.Name)
		}
		decl.Methods.Set(method.Name(), method)
	}
}
