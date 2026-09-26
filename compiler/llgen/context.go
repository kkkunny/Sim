package llgen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/go-llvm/ir"
	"github.com/kkkunny/go-llvm/target"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
)

// Ident 符号信息
type Ident struct {
	Name string

	ExternalFunc bool // 是外部函数
	FuncSymbol   bool // 是函数符号（全局函数/外部函数）：值表示为 {fnptr, null}，Local 为空

	Local llvm.Value[llvm.PtrT] // 局部变量/参数的存储地址
}

// Context 跨包共享的代码生成上下文
type Context struct {
	llvm       *llvm.Context
	target     *target.TargetMachine
	dataLayout *llvm.DataLayout

	triple string

	idents    map[hir.Ident]*Ident
	typeCache map[*globals.TypeDef]llvm.AnyType

	namedTypes map[string]llvm.StructType // named struct 按名去重（LLVM Context 级共享）

	eqFuncs     map[string]ir.Function // 相等性辅助函数缓存（键含模块标识与类型键，见 equal.go）
	moduleCount int                    // 已创建模块计数（辅助函数缓存的模块标识）

	closureCount int // 已创建闭包计数（ctx 结构体/包装函数命名，Context 级唯一避免跨模块同名）
}

// NewContext 创建代码生成上下文（初始化本机 LLVM 目标）
//
// 重定位模型用 PIC：clang 默认按 PIE 链接，内部/私有全局（如字符串常量）在静态
// 重定位模型下会被后端用 R_X86_64_32 绝对地址引用，链接时报
// "relocation ... can not be used when making a PIE object"，PIC 则生成 RIP 相对寻址。
func NewContext() *Context {
	stlerr.Must(target.InitNative())
	native := stlerr.MustWith(target.NativeTarget())
	tm := stlerr.MustWith(target.NewTargetMachine(
		native, target.DefaultTriple(), target.HostCPUName(), target.HostCPUFeatures(),
		target.OptNone, target.RelocPIC, target.CodeModelDefault,
	))
	return &Context{
		llvm:       llvm.NewContext(),
		target:     tm,
		dataLayout: tm.DataLayout(),

		triple: target.DefaultTriple(),

		idents:    make(map[hir.Ident]*Ident),
		typeCache: make(map[*globals.TypeDef]llvm.AnyType),

		namedTypes: make(map[string]llvm.StructType),

		eqFuncs: make(map[string]ir.Function),
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

// DataLayout 返回共享的目标数据布局（union payload 尺寸/对齐计算的权威来源）
func (c *Context) DataLayout() *llvm.DataLayout {
	return c.dataLayout
}

// NamedStruct 获取（或创建）名为 name 的 named struct；首次创建时 body 为空
func (c *Context) NamedStruct(name string) llvm.StructType {
	if t, ok := c.namedTypes[name]; ok {
		return t
	}
	t := c.llvm.NamedStruct(name)
	c.namedTypes[name] = t
	return t
}

// NamedStructWithBody 获取（或创建并填充）名为 name 的 named struct
func (c *Context) NamedStructWithBody(name string, elems ...llvm.AnyType) llvm.StructType {
	t := c.NamedStruct(name)
	if t.IsOpaque() {
		t.SetBody(elems, false)
	}
	return t
}

// Close 释放上下文资源
func (c *Context) Close() error {
	err1 := c.dataLayout.Close()
	err2 := c.target.Close()
	err3 := c.llvm.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}
