package analyse

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/mean"

	"github.com/kkkunny/Sim/ast"
)

// Analyser 语义分析器
type Analyser struct {
	parent *Analyser
	asts   linkedlist.LinkedList[ast.Global]

	pkgs       *hashmap.HashMap[string, *_PkgScope]
	pkgScope   *_PkgScope
	localScope _LocalScope

	selfValue *mean.Param
	selfType *mean.StructDef

	typeAliasTrace hashset.HashSet[*ast.TypeAlias]
}

func New(path string, asts linkedlist.LinkedList[ast.Global], target *llvm.Target) *Analyser {
	if target != nil {
		mean.Isize.Bits = target.PointerSize()
		mean.Usize.Bits = target.PointerSize()
	}
	pkgs := hashmap.NewHashMap[string, *_PkgScope]()
	return &Analyser{
		asts:     asts,
		pkgs:     &pkgs,
		pkgScope: _NewPkgScope(path),
		typeAliasTrace: hashset.NewHashSet[*ast.TypeAlias](),
	}
}

func newSon(parent *Analyser, path string, asts linkedlist.LinkedList[ast.Global]) *Analyser {
	return &Analyser{
		parent:   parent,
		asts:     asts,
		pkgs:     parent.pkgs,
		pkgScope: _NewPkgScope(path),
		typeAliasTrace: hashset.NewHashSet[*ast.TypeAlias](),
	}
}

func (self *Analyser) checkLoopImport(path string) bool {
	if self.pkgScope.path == path {
		return true
	}
	if self.parent != nil {
		return self.parent.checkLoopImport(path)
	}
	return false
}

// Analyse 分析语义
func (self *Analyser) Analyse() linkedlist.LinkedList[mean.Global] {
	meanNodes := linkedlist.NewLinkedList[mean.Global]()

	// 包
	if !self.pkgScope.IsBuildIn() {
		meanNodes.Append(self.importBuildInPackage())
	}
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		if im, ok := v.(*ast.Import); ok {
			meanNodes.Append(self.analyseImport(im))
		}
		return true
	})

	// 类型
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			self.declTypeDef(node)
		case *ast.TypeAlias:
			self.declTypeAlias(node)
		}
		return true
	})
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			meanNodes.PushBack(self.defTypeDef(node))
		case *ast.TypeAlias:
			self.defTypeAlias(node.Name.Source())
		}
		return true
	})

	// 值
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		if global := self.analyseGlobalDef(v); global != nil {
			meanNodes.PushBack(global)
		}
		return true
	})
	return meanNodes
}
