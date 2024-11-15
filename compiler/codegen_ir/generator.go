package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/queue"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	llvmUtil "github.com/kkkunny/Sim/compiler/codegen_ir/llvm"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	donePkgs set.Set[*global.Package]
	builder  *llvmUtil.Builder

	types  hashmap.HashMap[types.CustomType, llvm.Type]
	values hashmap.HashMap[values.Value, llvm.Value]

	funcCache        hashmap.HashMap[string, llvm.Function]
	loops            hashmap.HashMap[local.Loop, loop]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[values.Ident, llvm.Value]]
}

func New(target llvm.Target) *CodeGenerator {
	return &CodeGenerator{
		donePkgs:         set.StdHashSetWith[*global.Package](),
		builder:          llvmUtil.NewBuilder(target),
		types:            hashmap.AnyWith[types.CustomType, llvm.Type](),
		values:           hashmap.StdWith[values.Value, llvm.Value](),
		funcCache:        hashmap.StdWith[string, llvm.Function](),
		loops:            hashmap.StdWith[local.Loop, loop](),
		lambdaCaptureMap: queue.New[hashmap.HashMap[values.Ident, llvm.Value]](),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen(pkg *global.Package) llvm.Module {
	self.codegenPkg(pkg)
	// 初始化函数
	for _, b := range self.builder.GetInitFunction().Blocks() {
		if !b.IsTerminating() {
			self.builder.MoveToAfter(b)
			self.builder.CreateRet(nil)
		}
	}
	// 主函数
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		fIr, ok := iter.Value().(*global.FuncDef)
		if !ok || fIr.Name() != "main" {
			continue
		}
		f := self.values.Get(fIr).(llvm.Function)
		self.builder.MoveToAfter(stlslices.First(self.builder.GetMainFunction().Blocks()))
		self.builder.CreateCall("", f.FunctionType(), f)
		self.builder.CreateRet(stlval.Ptr[llvm.Value](self.builder.ConstInteger(self.builder.IntegerType(8), 0)))
		break
	}
	return self.builder.Module
}

func (self *CodeGenerator) codegenPkg(pkg *global.Package) {
	// 导入包
	self.codegenImportPkgs(pkg)
	// 类型声明
	self.codegenTypeDefDecl(pkg)
	// 类型定义
	self.codegenTypeDefDef(pkg)
	// 变量声明
	self.codegenGlobalVarDecl(pkg)
	// 变量定义
	self.codegenGlobalVarDef(pkg)
}
