package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/iterator"
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
	selfType  hir.Type

	typeAliasTrace hashset.HashSet[*ast.TypeAlias]
	genericFuncScope hashmap.HashMap[*hir.GenericFuncDef, *_FuncScope]
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
		typeAliasTrace: hashset.NewHashSet[*ast.TypeAlias](),
		genericFuncScope: hashmap.NewHashMap[*hir.GenericFuncDef, *_FuncScope](),
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
		typeAliasTrace: hashset.NewHashSet[*ast.TypeAlias](),
		genericFuncScope: hashmap.NewHashMap[*hir.GenericFuncDef, *_FuncScope](),
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
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if im, ok := v.(*ast.Import); ok {
			meanNodes.Append(self.analyseImport(im))
		}
		return true
	})

	// trait
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if trait, ok := v.(*ast.Trait); ok {
			self.declTrait(trait)
		}
		return true
	})

	// 类型
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			self.declStructDef(node)
		case *ast.TypeAlias:
			self.declTypeAlias(node)
		case *ast.GenericStructDef:
			self.declGenericStructDef(node)
		}
		return true
	})
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		switch node := v.(type) {
		case *ast.StructDef:
			meanNodes.PushBack(self.defStructDef(node))
		case *ast.TypeAlias:
			self.defTypeAlias(node.Name.Source())
		case *ast.GenericStructDef:
			meanNodes.PushBack(self.defGenericStructDef(node))
		}
		return true
	})
	// 类型循环检测
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		trace := hashset.NewHashSet[hir.Type]()
		var circle bool
		var name token.Token
		switch node := v.(type) {
		case *ast.StructDef:
			st, _ := self.pkgScope.GetStruct("", node.Name.Source())
			circle, name = self.checkTypeCircle(&trace, st), node.Name
		case *ast.TypeAlias:
			data, _ := self.pkgScope.GetTypeAlias("", node.Name.Source())
			alias, _ := data.Right()
			circle, name = self.checkTypeCircle(&trace, alias), node.Name
		}
		if circle{
			errors.ThrowCircularReference(name.Position, name)
		}
		return true
	})

	// 值
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	iterator.Foreach[ast.Global](self.asts, func(v ast.Global) bool {
		if global := self.analyseGlobalDef(v); global != nil {
			meanNodes.PushBack(global)
		}
		return true
	})
	return meanNodes
}

func (self *Analyser) checkTypeCircle(trace *hashset.HashSet[hir.Type], t hir.Type)bool{
	if trace.Contain(t){
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *hir.EmptyType, hir.NumberType, *hir.FuncType, *hir.BoolType, *hir.StringType:
	case *hir.PtrType:
		return self.checkTypeCircle(trace, typ.Elem)
	case *hir.RefType:
		return self.checkTypeCircle(trace, typ.Elem)
	case *hir.ArrayType:
		return self.checkTypeCircle(trace, typ.Elem)
	case *hir.TupleType:
		for _, e := range typ.Elems{
			if self.checkTypeCircle(trace, e){
				return true
			}
		}
	case *hir.StructType:
		for iter:=typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeCircle(trace, iter.Value().Second.Second){
				return true
			}
		}
	case *hir.UnionType:
		for _, e := range typ.Elems {
			if self.checkTypeCircle(trace, e){
				return true
			}
		}
	default:
		panic("unreachable")
	}
	return false
}
