package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/util"
)

func (self *CodeGenerator) buildEqual(t mean.Type, l, r mir.Value, not bool) mir.Value {
	switch meanType := t.(type) {
	case mean.IntType, *mean.BoolType, *mean.FloatType:
		return self.builder.BuildCmp(stlbasic.Ternary(!not, mir.CmpKindEQ, mir.CmpKindNE), l, r)
	case *mean.FuncType, *mean.PtrType, *mean.RefType:
		return self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), l, r)
	case *mean.ArrayType:
		at := self.codegenArrayType(meanType)
		lp, rp := self.builder.BuildAllocFromStack(at), self.builder.BuildAllocFromStack(at)
		self.builder.BuildStore(l, lp)
		self.builder.BuildStore(r, rp)

		res := self.buildArrayEqual(meanType, lp, rp)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	case *mean.TupleType, *mean.StructType:
		res := self.buildStructEqual(meanType, l, r)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	case *mean.StringType:
		name := "sim_runtime_str_eq_str"
		ft := self.ctx.NewFuncType(self.ctx.Bool(), l.Type(), r.Type())
		var f *mir.Function
		var ok bool
		if f, ok = self.module.NamedFunction(name); !ok {
			f = self.module.NewFunction(name, ft)
		}
		res := self.builder.BuildCall(f, l, r)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	case *mean.UnionType:
		// TODO: 联合类型比较
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayEqual(meanType *mean.ArrayType, l, r mir.Value) mir.Value {
	t := self.codegenArrayType(meanType)
	if t.Length() == 0 {
		return mir.Bool(self.ctx, true)
	}

	indexPtr := self.builder.BuildAllocFromStack(self.ctx.Usize())
	self.builder.BuildStore(mir.NewZero(indexPtr.ElemType()), indexPtr)
	condBlock := self.builder.CurrentBlock().Belong().NewBlock("")
	self.builder.CreateBr(condBlock)

	// cond
	self.builder.MoveToAfter(condBlock)
	var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(t.Capacity())))
	bodyBlock, outBlock := self.builder.CurrentBlock().Belong().NewBlock(""), self.builder.CurrentBlock().Belong().NewBlock("")
	self.builder.CreateCondBr(cond, bodyBlock, outBlock)

	// body
	self.builder.MoveToAfter(bodyBlock)
	index := self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr)
	cond = self.buildEqual(meanType.Elem, self.buildArrayIndex(t, l, index, true), self.buildArrayIndex(t, r, index, true), false)
	bodyEndBlock := self.builder.CurrentBlock()
	actionBlock := bodyEndBlock.Belong().NewBlock("")
	self.builder.CreateCondBr(cond, actionBlock, outBlock)

	// action
	self.builder.MoveToAfter(actionBlock)
	self.builder.CreateStore(self.builder.CreateUAdd("", index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
	self.builder.CreateBr(condBlock)

	// out
	self.builder.MoveToAfter(outBlock)
	phi := self.builder.CreatePHI("", self.ctx.IntegerType(1))
	phi.AddIncomings(
		struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.ctx.ConstInteger(self.ctx.IntegerType(1), 1), Block: condBlock},
		struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.ctx.ConstInteger(self.ctx.IntegerType(1), 0), Block: bodyEndBlock},
	)
	return phi
}

func (self *CodeGenerator) buildStructEqual(meanType mean.Type, l, r mir.Value) mir.Value {
	_, isTuple := meanType.(*mean.TupleType)
	t := stlbasic.TernaryAction(isTuple, func() llvm.StructType {
		return self.codegenTupleType(meanType.(*mean.TupleType))
	}, func() llvm.StructType {
		return self.codegenStructType(meanType.(*mean.StructType))
	})
	fields := stlbasic.TernaryAction(isTuple, func() dynarray.DynArray[mean.Type] {
		return dynarray.NewDynArrayWith(meanType.(*mean.TupleType).Elems...)
	}, func() dynarray.DynArray[mean.Type] {
		values := meanType.(*mean.StructType).Fields.Values()
		res := dynarray.NewDynArrayWithLength[mean.Type](values.Length())
		var i uint
		for iter:=values.Iterator(); iter.Next(); {
			res.Set(i, iter.Value().Second)
			i++
		}
		return res
	})

	if t.CountElems() == 0 {
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), 1)
	}

	beginBlock := self.builder.CurrentBlock()
	nextBlocks := dynarray.NewDynArrayWithLength[pair.Pair[llvm.Value, llvm.Block]](uint(t.CountElems()))
	srcBlocks := dynarray.NewDynArrayWithLength[llvm.Block](uint(t.CountElems()))
	for i := uint32(0); i < t.CountElems(); i++ {
		cond := self.buildEqual(fields.Get(uint(i)), self.buildStructIndex(t, l, uint(i), true), self.buildStructIndex(t, r, uint(i), true), false)
		nextBlock := beginBlock.Belong().NewBlock("")
		nextBlocks.Set(uint(i), pair.NewPair(cond, nextBlock))
		srcBlocks.Set(uint(i), self.builder.CurrentBlock())
		self.builder.MoveToAfter(nextBlock)
	}
	self.builder.MoveToAfter(beginBlock)

	endBlock := nextBlocks.Back().Second
	for iter := nextBlocks.Iterator(); iter.Next(); {
		item := iter.Value()
		if iter.HasNext() {
			self.builder.CreateCondBr(item.First, item.Second, endBlock)
		} else {
			self.builder.CreateBr(endBlock)
		}
		self.builder.MoveToAfter(item.Second)
	}

	phi := self.builder.CreatePHI("", self.ctx.IntegerType(1))
	for iter := srcBlocks.Iterator(); iter.Next(); {
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: stlbasic.Ternary[llvm.Value](iter.HasNext(), self.ctx.ConstInteger(self.ctx.IntegerType(1), 0), nextBlocks.Back().First), Block: iter.Value()})
	}
	return phi
}

func (self *CodeGenerator) getMainFunction() llvm.Function {
	mainFn, ok := self.module.GetFunction("main")
	if !ok {
		mainFn = self.module.NewFunction("main", self.ctx.FunctionType(false, self.ctx.IntegerType(8)))
		mainFn.NewBlock("entry")
	}
	return mainFn
}

func (self *CodeGenerator) getInitFunction() llvm.Function {
	initFn, ok := self.module.GetFunction("sim_runtime_init")
	if !ok {
		initFn = self.module.NewFunction("sim_runtime_init", self.ctx.FunctionType(false, self.ctx.VoidType()))
		self.module.AddConstructor(65535, initFn)
		initFn.NewBlock("entry")
	}
	return initFn
}

// NOTE: 跟系统、架构相关
func (self *CodeGenerator) enterFunction(ft llvm.FunctionType, f llvm.Function, paramNodes []*mean.Param) {
	switch {
	case self.target.IsWindows():
		var params []llvm.Param
		if stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.Array](ft.ReturnType()) {
			params = f.Params()[1:]
		} else {
			params = f.Params()
		}
		for i, p := range params {
			paramNode := paramNodes[i]

			pt := self.codegenType(paramNode.GetType())
			if stlbasic.Is[llvm.StructType](pt) || stlbasic.Is[llvm.Array](pt) {
				self.values[paramNode] = p
			} else {
				param := self.builder.CreateAlloca("", pt)
				self.builder.CreateStore(p, param)
				self.values[paramNode] = param
			}
		}
	case self.target.IsLinux():
		for i, p := range f.Params() {
			paramNode := paramNodes[i]
	
			pt := self.codegenType(paramNode.GetType())
			param := self.builder.CreateAlloca("", pt)
			self.builder.CreateStore(p, param)
			self.values[paramNode] = param
		}
	default:
		panic("unreachable")
	}
}

// CodegenIr 中间代码生成
func CodegenIr(path string) (llvm.Module, stlerror.Error) {
	means, err := analyse.Analyse(path)
	if err != nil{
		return llvm.Module{}, err
	}
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	if err = stlerror.ErrorWrap(llvm.InitializeNativeTarget()); err != nil{
		return llvm.Module{}, err
	}
	target, err := stlerror.ErrorWith(llvm.NativeTarget())
	if err != nil{
		return llvm.Module{}, err
	}
	module := New(target, means).Codegen()
	if err = stlerror.ErrorWrap(module.Verify()); err != nil{
		return llvm.Module{}, err
	}
	return module, nil
}
