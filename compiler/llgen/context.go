package llgen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir"
)

type Ident struct {
	Name string
}

// Context 跨包共享的LLVM上下文与符号信息
type Context struct {
	llvmCtx *llvm.Context

	idents       map[hir.Ident]*Ident
	typeCache    map[string]llvm.AnyType
	definedTypes set.Set[string]

	ptrType     llvm.PtrType
	i8Type      llvm.IntType
	i32Type     llvm.IntType
	i64Type     llvm.IntType
	emptyType   llvm.StructType
	strType     llvm.StructType
	funcPair    llvm.StructType
	fatFuncType llvm.StructType // 胖函数，用于变量赋值
}

func NewContext() *Context {
	ctx := &Context{
		llvmCtx: llvm.NewContext(),

		idents:       make(map[hir.Ident]*Ident),
		typeCache:    make(map[string]llvm.AnyType),
		definedTypes: set.AnyHashSetWith[string](),
	}

	llctx := ctx.llvmCtx
	ctx.ptrType = llctx.Ptr(0)
	ctx.i8Type = llctx.Int(8)
	ctx.i32Type = llctx.Int(32)
	ctx.i64Type = llctx.Int(64)
	ctx.emptyType = llctx.Struct(nil, false)
	ctx.strType = llctx.Struct([]llvm.AnyType{ctx.ptrType, ctx.i64Type}, false)
	ctx.funcPair = llctx.Struct([]llvm.AnyType{ctx.ptrType, ctx.ptrType}, false)
	ctx.fatFuncType = llctx.Struct([]llvm.AnyType{ctx.funcPair, ctx.ptrType}, false)
	return ctx
}

func (ctx *Context) Close() error {
	return ctx.llvmCtx.Close()
}
