package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/stack"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/global"
)

// Analyser 语义分析器
type Analyser struct {
	importStack stack.Stack[*global.Package] // 包的依赖链，不包括当前包
	allPkgs     hashmap.HashMap[stlos.FilePath, *global.Package]
	pkg         *global.Package
}

func (self *Analyser) Analyse(asts linkedlist.LinkedList[ast.Ast]) *global.Package {
	// 类型声明
	self.analyseTypeDecl(asts)

	// 类型定义
	self.analyseTypeDef(asts)

	// 变量声明

	// 变量定义

	return self.pkg
}

// 类型声明
func (self *Analyser) analyseTypeDecl(asts linkedlist.LinkedList[ast.Ast]) {
	for iter := asts.Iterator(); iter.HasNext(); {
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
func (self *Analyser) analyseTypeDef(asts linkedlist.LinkedList[ast.Ast]) {
	for iter := asts.Iterator(); iter.HasNext(); {
		switch node := iter.Value().(type) {
		case *ast.Trait:
			self.defTrait(node)
		case *ast.TypeDef:
			self.defTypeDef(node)
		case *ast.TypeAlias:
			self.defTypeAlias(node)
		}
	}
}
