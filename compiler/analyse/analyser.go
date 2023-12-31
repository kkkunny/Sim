package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"

	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/token"

	"github.com/kkkunny/Sim/ast"
)

// Analyser 语义分析器
type Analyser struct {
	parent *Analyser
	asts   linkedlist.LinkedList[ast.Global]

	pkgs       *hashmap.HashMap[hir.Package, *_PkgScope]
	pkgScope   *_PkgScope
	localScope _LocalScope

	selfValue *hir.Param
	selfType  hir.TypeDef

	typeAliasTrace hashset.HashSet[*ast.TypeAlias]
	genericIdentMap hashmap.HashMap[string, *hir.GenericIdentType]
}

func New(asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg hir.Package
	if !asts.Empty(){
		pkg = stlerror.MustWith(hir.NewPackage(asts.Front().Position().Reader.Path().Dir()))
	}
	pkgs := hashmap.NewHashMap[hir.Package, *_PkgScope]()
	return &Analyser{
		asts:     asts,
		pkgs:     &pkgs,
		pkgScope: _NewPkgScope(pkg),
	}
}

func newSon(parent *Analyser, asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg hir.Package
	if !asts.Empty(){
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
func (self *Analyser) Analyse() linkedlist.LinkedList[hir.Global] {
	meanNodes := linkedlist.NewLinkedList[hir.Global]()

	// 包
	if self.pkgScope.pkg != hir.BuildInPackage {
		hirs, _ := self.importPackage(hir.BuildInPackage, "", true)
		meanNodes.Append(hirs)
	}
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if im, ok := v.(*ast.Import); ok {
			meanNodes.Append(self.analyseImport(im))
		}
		return true
	})

	// 类型
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			if node.Name.Params.IsNone(){
				self.declStructDef(node)
			}else{
				self.declGenericStructDef(node)
			}
		case *ast.TypeAlias:
			self.declTypeAlias(node)
		}
		return true
	})
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			if node.Name.Params.IsNone(){
				meanNodes.PushBack(self.defStructDef(node))
			}else{
				meanNodes.PushBack(self.defGenericStructDef(node))
			}
		case *ast.TypeAlias:
			meanNodes.PushBack(self.defTypeAlias(node))
		}
		return true
	})
	// 类型循环检测
	stliter.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		trace := hashset.NewHashSet[hir.Type]()
		var circle bool
		var name token.Token
		switch node := v.(type) {
		case *ast.StructDef:
			if node.Name.Params.IsNone(){
				st, _ := self.pkgScope.getLocalTypeDef(node.Name.Name.Source())
				circle, name = self.checkTypeCircle(&trace, st), node.Name.Name
			}
		case *ast.TypeAlias:
			tad, _ := self.pkgScope.getLocalTypeDef(node.Name.Source())
			circle, name = self.checkTypeCircle(&trace, tad), node.Name
		}
		if circle{
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
			meanNodes.PushBack(global)
		}
		return true
	})
	return meanNodes
}
