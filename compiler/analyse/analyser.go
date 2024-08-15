package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
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

	pkgs       *hashmap.HashMap[oldhir.Package, *_PkgScope]
	pkgScope   *_PkgScope
	localScope _LocalScope

	selfCanBeNil bool
	selfType     *oldhir.CustomType

	typeAliasTrace hashset.HashSet[*ast.TypeAlias]
}

func New(asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg oldhir.Package
	if !asts.Empty() {
		pkg = stlerror.MustWith(oldhir.NewPackage(asts.Front().Position().Reader.Path().Dir()))
	}
	pkgs := hashmap.NewHashMap[oldhir.Package, *_PkgScope]()
	return &Analyser{
		asts:     asts,
		pkgs:     &pkgs,
		pkgScope: _NewPkgScope(pkg),
	}
}

func newSon(parent *Analyser, asts linkedlist.LinkedList[ast.Global]) *Analyser {
	var pkg oldhir.Package
	if !asts.Empty() {
		pkg = stlerror.MustWith(oldhir.NewPackage(asts.Front().Position().Reader.Path().Dir()))
	}
	return &Analyser{
		parent:   parent,
		asts:     asts,
		pkgs:     parent.pkgs,
		pkgScope: _NewPkgScope(pkg),
	}
}

// Analyse 分析语义
func (self *Analyser) Analyse() *oldhir.Result {
	globalIrs := linkedlist.NewLinkedList[oldhir.Global]()

	// 包
	if self.pkgScope.pkg != oldhir.BuildInPackage {
		hirs, _ := self.importPackage(oldhir.BuildInPackage, "", true)
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
		trace := hashset.NewHashSet[oldhir.Type]()
		var circle bool
		var name token.Token
		switch node := v.(type) {
		case *ast.TypeDef:
			st, _ := self.pkgScope.getLocalTypeDef(node.Name.Source())
			circle, name = self.checkTypeDefCircle(&trace, st), node.Name
		case *ast.TypeAlias:
			tad, _ := self.pkgScope.getLocalTypeDef(node.Name.Source())
			circle, name = self.checkTypeAliasCircle(&trace, tad), node.Name
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
	return &oldhir.Result{
		Globals: globalIrs,
		BuildinTypes: struct {
			Isize   oldhir.Type
			I8      oldhir.Type
			I16     oldhir.Type
			I32     oldhir.Type
			I64     oldhir.Type
			Usize   oldhir.Type
			U8      oldhir.Type
			U16     oldhir.Type
			U32     oldhir.Type
			U64     oldhir.Type
			F32     oldhir.Type
			F64     oldhir.Type
			Bool    oldhir.Type
			Str     oldhir.Type
			Default *oldhir.Trait
			Copy    *oldhir.Trait
			Add     *oldhir.Trait
			Sub     *oldhir.Trait
			Mul     *oldhir.Trait
			Div     *oldhir.Trait
			Rem     *oldhir.Trait
			And     *oldhir.Trait
			Or      *oldhir.Trait
			Xor     *oldhir.Trait
			Shl     *oldhir.Trait
			Shr     *oldhir.Trait
			Eq      *oldhir.Trait
			Lt      *oldhir.Trait
			Gt      *oldhir.Trait
			Land    *oldhir.Trait
			Lor     *oldhir.Trait
			Neg     *oldhir.Trait
			Not     *oldhir.Trait
		}{
			Isize:   self.pkgScope.Isize(),
			I8:      self.pkgScope.I8(),
			I16:     self.pkgScope.I16(),
			I32:     self.pkgScope.I32(),
			I64:     self.pkgScope.I64(),
			Usize:   self.pkgScope.Usize(),
			U8:      self.pkgScope.U8(),
			U16:     self.pkgScope.U16(),
			U32:     self.pkgScope.U32(),
			U64:     self.pkgScope.U64(),
			F32:     self.pkgScope.F32(),
			F64:     self.pkgScope.F64(),
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
