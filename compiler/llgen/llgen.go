package llgen

import (
	"fmt"

	"github.com/kkkunny/go-llvm/ir"

	"github.com/kkkunny/Sim/compiler/hir/globals"
)

// CodeGenerator HIR → LLVM IR 代码生成器（每个包一个实例、一个 LLVM 模块）
type CodeGenerator struct {
	pkg     *globals.Package
	ctx     *Context
	module  *ir.Module
	builder *ir.Builder
}

// New 创建某包的代码生成器
func New(ctx *Context, pkg *globals.Package) *CodeGenerator {
	module := ir.NewModule(ctx.LLVM(), pkg.Name)
	ctx.TargetMachine().ApplyTo(module)
	return &CodeGenerator{
		pkg:     pkg,
		ctx:     ctx,
		module:  module,
		builder: ir.NewBuilder(ctx.LLVM()),
	}
}

// Module 返回当前包对应的 LLVM 模块
func (c *CodeGenerator) Module() *ir.Module {
	return c.module
}

// Close 释放当前生成器资源
func (c *CodeGenerator) Close() error {
	err1 := c.builder.Close()
	err2 := c.module.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Generate 生成包对应的 LLVM 模块
func (c *CodeGenerator) Generate() *ir.Module {
	for _, g := range c.pkg.Globals {
		c.genTypeDecl(g)
	}
	for _, g := range c.pkg.Globals {
		c.genTypeDef(g)
	}
	for _, g := range c.pkg.Globals {
		c.genGlobalValue(g)
	}

	if err := c.module.Verify(); err != nil {
		panic(fmt.Errorf("包 %s 生成 LLVM IR 校验失败: %w", c.pkg.Path, err))
	}
	return c.module
}
