package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/set"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/hir"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"

	"github.com/kkkunny/Sim/compiler/ast"
)

// Analyser 语义分析器
type Analyser struct {
	parent *Analyser
	asts   linkedlist.LinkedList[ast.Global]

	pkgs       hashmap.HashMap[hir.Package, *_PkgScope]
	pkgScope   *_PkgScope
	localScope _LocalScope

	selfCanBeNil bool
	selfType     *hir.CustomType

	typeAliasTrace set.Set[*ast.TypeAlias]
}

func New(asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg hir.Package
	if !asts.Empty() {
		pkg = stlerror.MustWith(hir.NewPackage(asts.Front().Position().Reader.Path().Dir()))
	}
	return &Analyser{
		asts:           asts,
		pkgs:           hashmap.StdWith[hir.Package, *_PkgScope](),
		pkgScope:       _NewPkgScope(pkg),
		typeAliasTrace: set.StdHashSetWith[*ast.TypeAlias](),
	}
}

func newSon(parent *Analyser, asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg hir.Package
	if !asts.Empty() {
		pkg = stlerror.MustWith(hir.NewPackage(asts.Front().Position().Reader.Path().Dir()))
	}
	return &Analyser{
		parent:   parent,
		asts:     asts,
		pkgs:     parent.pkgs,
		pkgScope: _NewPkgScope(pkg),
	}
}

// Analyse 分析语义
func (self *Analyser) Analyse() *hir.Result {
	globalIrs := linkedlist.NewLinkedList[hir.Global]()

	// 包
	if self.pkgScope.pkg != hir.BuildInPackage {
		hirs, _ := self.importPackage(hir.BuildInPackage, "", true)
		globalIrs.Append(hirs)
	}
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if im, ok := v.(*ast.Import); ok {
			globalIrs.Append(self.analyseImport(im))
		}
		return true
	})

	// 类型
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.TypeDef:
			self.declTypeDef(node)
		case *ast.TypeAlias:
			self.declTypeAlias(node)
		case *ast.Trait:
			self.declTrait(node)
		}
		return true
	})
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.TypeDef:
			globalIrs.PushBack(self.defTypeDef(node))
		case *ast.TypeAlias:
			globalIrs.PushBack(self.defTypeAlias(node))
		case *ast.Trait:
			globalIrs.PushBack(self.defTrait(node))
		}
		return true
	})
	// 类型循环检测
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		trace := set.StdHashSetWith[hir.Type]()
		var circle bool
		var name token.Token
		switch node := v.(type) {
		case *ast.TypeDef:
			st, _ := self.pkgScope.getLocalTypeDef(node.Name.Source())
			circle, name = self.checkTypeDefCircle(trace, st), node.Name
		case *ast.TypeAlias:
			tad, _ := self.pkgScope.getLocalTypeDef(node.Name.Source())
			circle, name = self.checkTypeAliasCircle(trace, tad), node.Name
		}
		if circle {
			errors.ThrowCircularReference(name.Position, name)
		}
		return true
	})

	// 值
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if global := self.analyseGlobalDef(v); global != nil {
			globalIrs.PushBack(global)
		}
		return true
	})
	return &hir.Result{
		Globals: globalIrs,
		BuildinTypes: struct {
			Bool    hir.Type
			Str     hir.Type
			Default *hir.Trait
			Copy    *hir.Trait
			Add     *hir.Trait
			Sub     *hir.Trait
			Mul     *hir.Trait
			Div     *hir.Trait
			Rem     *hir.Trait
			And     *hir.Trait
			Or      *hir.Trait
			Xor     *hir.Trait
			Shl     *hir.Trait
			Shr     *hir.Trait
			Eq      *hir.Trait
			Lt      *hir.Trait
			Gt      *hir.Trait
			Land    *hir.Trait
			Lor     *hir.Trait
			Neg     *hir.Trait
			Not     *hir.Trait
		}{
			Bool:    self.pkgScope.Bool(),
			Str:     self.pkgScope.Str(),
			Default: self.pkgScope.Default(),
			Copy:    self.pkgScope.Copy(),
			Add:     self.pkgScope.Add(),
			Sub:     self.pkgScope.Sub(),
			Mul:     self.pkgScope.Mul(),
			Div:     self.pkgScope.Div(),
			Rem:     self.pkgScope.Rem(),
			And:     self.pkgScope.And(),
			Or:      self.pkgScope.Or(),
			Xor:     self.pkgScope.Xor(),
			Shl:     self.pkgScope.Shl(),
			Shr:     self.pkgScope.Shr(),
			Eq:      self.pkgScope.Shr(),
			Lt:      self.pkgScope.Shr(),
			Gt:      self.pkgScope.Shr(),
			Land:    self.pkgScope.Shr(),
			Lor:     self.pkgScope.Shr(),
			Neg:     self.pkgScope.Neg(),
			Not:     self.pkgScope.Not(),
		},
	}
}
