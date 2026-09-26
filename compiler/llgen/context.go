package llgen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/target"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/hir"
)

// Ident 符号信息
type Ident struct {
	Name string

	ExternalFunc bool // 是外部函数
}

// Context 跨包共享的代码生成上下文
type Context struct {
	llvm   *llvm.Context
	target *target.TargetMachine

	triple string

	idents    map[hir.Ident]*Ident
	typeCache map[string]llvm.AnyType
}

// NewContext 创建代码生成上下文（初始化本机 LLVM 目标）
func NewContext() *Context {
	stlerr.Must(target.InitNative())
	native := stlerr.MustWith(target.NativeTarget())
	tm := stlerr.MustWith(target.NewTargetMachine(
		native, target.DefaultTriple(), target.HostCPUName(), target.HostCPUFeatures(),
		target.OptNone, target.RelocDefault, target.CodeModelDefault,
	))
	return &Context{
		llvm:   llvm.NewContext(),
		target: tm,

		triple: target.DefaultTriple(),

		idents:    make(map[hir.Ident]*Ident),
		typeCache: make(map[string]llvm.AnyType),
	}
}

// LLVM 返回底层 LLVM 上下文
func (c *Context) LLVM() *llvm.Context {
	return c.llvm
}

// TargetMachine 返回本机目标机器
func (c *Context) TargetMachine() *target.TargetMachine {
	return c.target
}

// Close 释放上下文资源
func (c *Context) Close() error {
	err1 := c.target.Close()
	err2 := c.llvm.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
