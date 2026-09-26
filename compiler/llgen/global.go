package llgen

import (
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
)

func (c *CodeGenerator) genTypeDecl(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		c.genCustomTypeDecl(global)
	}
}

func (c *CodeGenerator) genCustomTypeDecl(global *globals.TypeDef) {
	// TODO(B6/B7): 聚合类型预声明 opaque named struct
}

func (c *CodeGenerator) genTypeDef(global globals.Global) {
	switch global := global.(type) {
	case *globals.TypeDef:
		c.genCustomTypeDef(global)
	}
}

func (c *CodeGenerator) genCustomTypeDef(global *globals.TypeDef) {
	// TODO(B7): 自定义类型定义
}

func (c *CodeGenerator) genGlobalValue(global globals.Global) {
	switch global := global.(type) {
	case *locals.Let:
		c.genGlobalLet(global)
	}
}

func (c *CodeGenerator) genGlobalLet(l *locals.Let) {
	// TODO(C4/C5/C6): 全局变量、函数定义/声明、main 入口
}
