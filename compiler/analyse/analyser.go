package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/set"
	"github.com/kkkunny/stl/container/stack"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/token"
)

// Analyser 语义分析器
type Analyser struct {
	importStack stack.Stack[*global.Package]                     // 包的依赖链，从main包开始，不包括当前包
	allPkgs     hashmap.HashMap[stlos.FilePath, *global.Package] // main包的所有依赖包
	pkg         *global.Package
	scope       local.Scope // 当前所处作用域
}

func New(path stlos.FilePath) *Analyser {
	return &Analyser{
		importStack: stack.New[*global.Package](),
		allPkgs:     hashmap.StdWith[stlos.FilePath, *global.Package](),
		pkg:         global.NewPackage(path),
	}
}

func newSon(parent *Analyser, pkg *global.Package) *Analyser {
	return &Analyser{
		importStack: parent.importStack,
		allPkgs:     parent.allPkgs,
		pkg:         pkg,
	}
}

func (self *Analyser) Analyse(asts linkedlist.LinkedList[ast.Global]) *global.Package {
	// 包导入
	self.analyseSonPackage(asts)

	// 设置当前作用域
	self.scope = self.pkg

	// 类型声明
	self.analyseTypeDecl(asts)

	// 类型定义
	self.analyseTypeDef(asts)

	// 值声明
	self.analyseValueDecl(asts)

	// 值定义
	self.analyseValueDef(asts)

	return self.pkg
}

// 类型声明
func (self *Analyser) analyseSonPackage(asts linkedlist.LinkedList[ast.Global]) {
	if !self.pkg.IsBuildIn() {
		self.importPackage(config.BuildInPkgPath, "", true)
	}
	for iter := asts.Iterator(); iter.Next(); {
		switch node := iter.Value().(type) {
		case *ast.Import:
			self.analyseImport(node)
		}
	}
}

// 类型声明
func (self *Analyser) analyseTypeDecl(asts linkedlist.LinkedList[ast.Global]) {
	for iter := asts.Iterator(); iter.Next(); {
		switch node := iter.Value().(type) {
		case *ast.Trait:
			self.declTrait(node)
		case *ast.TypeDef:
			self.declTypeDef(node)
		case *ast.TypeAlias:
			self.declTypeAlias(node)
		}
	}
}

// 类型定义
func (self *Analyser) analyseTypeDef(asts linkedlist.LinkedList[ast.Global]) {
	typedefs := linkedhashmap.StdWith[global.TypeDef, token.Token]()
	for iter := asts.Iterator(); iter.Next(); {
		switch node := iter.Value().(type) {
		case *ast.Trait:
			self.defTrait(node)
		case *ast.TypeDef:
			typedefs.Set(self.defTypeDef(node), node.Name)
		case *ast.TypeAlias:
			typedefs.Set(self.defTypeAlias(node), node.Name)
		}
	}
	trace := set.AnyHashSetWith[types.Type]()
	for iter := typedefs.Iterator(); iter.Next(); {
		trace.Clear()
		circle := self.checkTypeCircle(trace, iter.Value().E1())
		if circle {
			errors.ThrowCircularReference(iter.Value().E2().Position, iter.Value().E2())
		}
	}
}

// 值声明
func (self *Analyser) analyseValueDecl(asts linkedlist.LinkedList[ast.Global]) {
	for iter := asts.Iterator(); iter.Next(); {
		switch node := iter.Value().(type) {
		case *ast.FuncDef:
			if node.SelfType.IsNone() {
				self.declFuncDef(node)
			} else {
				self.declMethodDef(node)
			}
		case *ast.SingleVariableDef:
			self.declSingleGlobalVariable(node)
		case *ast.MultipleVariableDef:
			self.declMultiGlobalVariable(node)
		}
	}
}

// 值定义
func (self *Analyser) analyseValueDef(asts linkedlist.LinkedList[ast.Global]) {
	for iter := asts.Iterator(); iter.Next(); {
		switch node := iter.Value().(type) {
		case *ast.FuncDef:
			if node.SelfType.IsNone() {
				self.defFuncDef(node)
			} else {
				self.defMethodDef(node)
			}
		case *ast.SingleVariableDef:
			self.defSingleGlobalVariable(node)
		case *ast.MultipleVariableDef:
			self.defMultiGlobalVariable(node)
		}
	}
}
