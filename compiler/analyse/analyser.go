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
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/token"
)

// Analyser 语义分析器
type Analyser struct {
	importStack stack.Stack[*hir.Package]                     // 包的依赖链，从main包开始，不包括当前包
	allPkgs     hashmap.HashMap[stlos.FilePath, *hir.Package] // main包的所有依赖包
	allFiles    hashmap.HashMap[stlos.FilePath, *hir.File]    // 文件路径与文件作用域映射
	pkg         *hir.Package                                  // 当前包
	scope       local.Scope                                   // 当前所处作用域
}

func New(path stlos.FilePath) *Analyser {
	return &Analyser{
		importStack: stack.New[*hir.Package](),
		allPkgs:     hashmap.StdWith[stlos.FilePath, *hir.Package](),
		allFiles:    hashmap.StdWith[stlos.FilePath, *hir.File](),
		pkg:         hir.NewPackage(path),
	}
}

func newSon(parent *Analyser, pkg *hir.Package) *Analyser {
	return &Analyser{
		importStack: parent.importStack,
		allPkgs:     parent.allPkgs,
		allFiles:    parent.allFiles,
		pkg:         pkg,
	}
}

func (self *Analyser) Analyse(astsList ...linkedlist.LinkedList[ast.Global]) *hir.Package {
	if len(astsList) == 0 {
		return self.pkg
	}

	self.scope = self.pkg
	if self.allPkgs.Contain(self.pkg.Path()) {
		self.allPkgs.Set(self.pkg.Path(), self.pkg)
	}

	// 包导入
	for _, asts := range astsList {
		self.analyseDepPackage(asts)
	}

	// 类型声明
	for _, asts := range astsList {
		self.analyseFileTypeDecl(asts)
	}

	// 类型定义
	for _, asts := range astsList {
		self.analyseFileTypeDef(asts)
	}

	// 值声明
	for _, asts := range astsList {
		self.analyseFileValueDecl(asts)
	}

	// 值定义
	for _, asts := range astsList {
		self.analyseFileValueDef(asts)
	}

	return self.pkg
}

func (self *Analyser) analyseFileTypeDecl(asts linkedlist.LinkedList[ast.Global]) {
	if asts.Empty() {
		return
	}

	self.scope = self.getFileByPath(asts.Front().Position())

	// 类型声明
	self.analyseTypeDecl(asts)
}

func (self *Analyser) analyseFileTypeDef(asts linkedlist.LinkedList[ast.Global]) {
	if asts.Empty() {
		return
	}

	self.scope = self.getFileByPath(asts.Front().Position())

	// 类型定义
	self.analyseTypeDef(asts)
}

func (self *Analyser) analyseFileValueDecl(asts linkedlist.LinkedList[ast.Global]) {
	if asts.Empty() {
		return
	}

	self.scope = self.getFileByPath(asts.Front().Position())

	// 值声明
	self.analyseValueDecl(asts)
}

func (self *Analyser) analyseFileValueDef(asts linkedlist.LinkedList[ast.Global]) {
	if asts.Empty() {
		return
	}

	self.scope = self.getFileByPath(asts.Front().Position())

	// 值定义
	self.analyseValueDef(asts)
}

// 依赖包
func (self *Analyser) analyseDepPackage(asts linkedlist.LinkedList[ast.Global]) {
	if asts.Empty() {
		return
	}

	filePath := asts.Front().Position().Reader.Path()
	file := hir.NewFile(filePath, self.pkg)
	self.allFiles.Set(filePath, file)
	self.pkg.AddFile(file)

	if !self.getFileByPath(asts.Front().Position()).Package().IsBuildIn() {
		_, _ = self.importPackage(asts.Front().Position(), config.BuildInPkgPath, "", true)
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
	typedefs := linkedhashmap.StdWith[types.TypeDef, token.Token]()
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
	trace := set.AnyHashSetWith[hir.Type]()
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
