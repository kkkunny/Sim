package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/queue"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	llvmUtil "github.com/kkkunny/Sim/compiler/codegen_ir/llvm"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	builder *llvmUtil.Builder
	pkg     *hir.Package

	values hashmap.HashMap[hir.Value, llvm.Value]

	funcCache        hashmap.HashMap[string, llvm.Function]
	loops            hashmap.HashMap[local.Loop, loop]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[values.Ident, llvm.Value]]
	virtualTypes     hashmap.HashMap[types.VirtualType, hir.Type]
}

func New(target llvm.Target, pkg *hir.Package) *CodeGenerator {
	return &CodeGenerator{
		builder:          llvmUtil.NewBuilder(llvm.GlobalContext, target),
		pkg:              pkg,
		values:           hashmap.StdWith[hir.Value, llvm.Value](),
		funcCache:        hashmap.StdWith[string, llvm.Function](),
		loops:            hashmap.StdWith[local.Loop, loop](),
		lambdaCaptureMap: queue.New[hashmap.HashMap[values.Ident, llvm.Value]](),
		virtualTypes:     hashmap.AnyWith[types.VirtualType, hir.Type](),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() llvm.Module {
	defer self.builder.Builder.Free()

	self.codegenPkg(self.pkg)

	// 初始化函数
	if _, ok := self.builder.GetFunction("sim_runtime_init"); ok {
		for _, b := range self.builder.GetInitFunction().Blocks() {
			if !b.IsTerminating() {
				self.builder.MoveToAfter(b)
				self.builder.CreateRet(nil)
			}
		}
	}
	// 主函数
	for iter := self.pkg.Globals().Iterator(); iter.Next(); {
		fIr, ok := iter.Value().(*global.FuncDef)
		if !ok || stlval.IgnoreWith(fIr.GetName()) != "main" {
			continue
		}
		f := self.values.Get(fIr).(llvm.Function)
		self.builder.MoveToAfter(stlslices.First(self.builder.GetMainFunction().Blocks()))
		self.builder.CreateCall("", f.FunctionType(), f)
		self.builder.CreateRet(self.builder.ConstInteger(self.builder.IntegerType(8), 0))
		break
	}
	return self.builder.Module
}

func (self *CodeGenerator) codegenPkg(pkg *hir.Package) {
	// 变量声明
	self.codegenGlobalVarDecl(pkg)
	// 变量定义
	self.codegenGlobalVarDef(pkg)
}
