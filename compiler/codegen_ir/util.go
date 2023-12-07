package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *CodeGenerator) buildArrayIndex(t llvm.ArrayType, from, index llvm.Value, load bool) llvm.Value {
	// NOTE: from必须是该数组的指针
	elemPtr := self.builder.CreateInBoundsGEP("", t, from, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
	if !load {
		return elemPtr
	}
	return self.builder.CreateLoad("", t.Element(), elemPtr)
}

func (self *CodeGenerator) buildArrayIndexWith(t llvm.ArrayType, from llvm.Value, index uint, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		return self.buildArrayIndex(t, from, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(index)), load)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractValue("", from, index)
	}
}

func (self *CodeGenerator) buildStructIndex(t llvm.StructType, from llvm.Value, index uint, load bool) llvm.Value {
	// TODO: 结构体和数组取下标和指针类型冲突
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		elemPtr := self.builder.CreateStructGEP("", t, from, index)
		if !load {
			return elemPtr
		}
		return self.builder.CreateLoad("", t.GetElem(uint32(index)), elemPtr)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractValue("", from, index)
	}
}

func (self *CodeGenerator) buildEqual(t mean.Type, l, r llvm.Value, not bool) llvm.Value {
	switch meanType := t.(type) {
	case *mean.EmptyType:
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), stlbasic.Ternary[int64](!not, 1, 0))
	case mean.IntType, *mean.BoolType, *mean.FuncType, *mean.PtrType, *mean.RefType:
		return self.builder.CreateIntCmp("", stlbasic.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
	case *mean.FloatType:
		return self.builder.CreateFloatCmp("", stlbasic.Ternary(!not, llvm.FloatOEQ, llvm.FloatUNE), l, r)
	case *mean.ArrayType:
		at := self.codegenArrayType(meanType)
		lp, rp := self.builder.CreateAlloca("", at), self.builder.CreateAlloca("", at)
		self.builder.CreateStore(l, lp)
		self.builder.CreateStore(r, rp)

		res := self.buildArrayEqual(meanType, lp, rp)
		if not {
			res = self.builder.CreateNot("", res)
		}
		return res
	case *mean.TupleType, *mean.StructType:
		res := self.buildStructEqual(meanType, l, r)
		if not {
			res = self.builder.CreateNot("", res)
		}
		return res
	case *mean.StringType:
		name := "sim_runtime_str_eq_str"
		ft := self.ctx.FunctionType(false, self.ctx.IntegerType(1), l.Type(), r.Type())
		var f llvm.Function
		var ok bool
		if f, ok = self.module.GetFunction(name); !ok {
			f = self.newFunction(name, ft)
		}
		res := self.buildCall(true, ft, f, l, r)
		if not {
			res = self.builder.CreateNot("", res)
		}
		return res
	case *mean.UnionType:
		// TODO: 联合类型比较
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayEqual(meanType *mean.ArrayType, l, r llvm.Value) llvm.Value {
	t := self.codegenArrayType(meanType)
	if t.Capacity() == 0 {
		return self.ctx.ConstInteger(self.ctx.IntegerType(1), 1)
	}

	indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
	self.builder.CreateStore(self.ctx.ConstNull(self.ctx.IntPtrType(self.target)), indexPtr)
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

func (self *CodeGenerator) buildStructEqual(meanType mean.Type, l, r llvm.Value) llvm.Value {
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
func (self *CodeGenerator) buildCall(load bool, ft llvm.FunctionType, f llvm.Value, param ...llvm.Value) llvm.Value {
	switch {
	case self.target.IsWindows():
		var retChangeToPtr bool
		retType, paramTypes := ft.ReturnType(), ft.Params()
		if stlbasic.Is[llvm.StructType](retType) || stlbasic.Is[llvm.Array](retType) {
			ptr := self.builder.CreateAlloca("", retType)
			param = append([]llvm.Value{ptr}, param...)
			paramTypes = append([]llvm.Type{self.ctx.PointerType(retType)}, paramTypes...)
			retType = self.ctx.VoidType()
			retChangeToPtr = true
		}
		for i, p := range param {
			if pt := p.Type(); stlbasic.Is[llvm.StructType](pt) || stlbasic.Is[llvm.Array](pt) {
				ptr := self.builder.CreateAlloca("", pt)
				self.builder.CreateStore(p, ptr)
				param[i] = ptr
				paramTypes[i] = self.ctx.PointerType(pt)
			}
		}
		var ret llvm.Value = self.builder.CreateCall("", self.ctx.FunctionType(false, retType, paramTypes...), f, param...)
		if retChangeToPtr {
			ret = param[0]
			if load {
				ret = self.builder.CreateLoad("", ft.ReturnType(), ret)
			}
		}
		return ret
	case self.target.IsLinux():
		return self.builder.CreateCall("", self.ctx.FunctionType(false, ft.ReturnType(), ft.Params()...), f, param...)
	default:
		panic("unreachable")
	}
}

// NOTE: 跟系统、架构相关
func (self *CodeGenerator) buildRet(ft llvm.FunctionType, ret *llvm.Value) {
	switch {
	case self.target.IsWindows():
		f := self.builder.CurrentBlock().Belong()

		retChangeToPtr := stlbasic.Is[llvm.StructType](ft.ReturnType()) || stlbasic.Is[llvm.Array](ft.ReturnType())
	
		if retChangeToPtr {
			self.builder.CreateStore(*ret, f.GetParam(0))
			self.builder.CreateRet(nil)
		} else {
			self.builder.CreateRet(ret)
		}
	case self.target.IsLinux():
		self.builder.CreateRet(ret)
	default:
		panic("unreachable")
	}
}

// NOTE: 跟系统、架构相关
func (self *CodeGenerator) newFunction(name string, t llvm.FunctionType) llvm.Function {
	switch {
	case self.target.IsWindows():
		ret, param := t.ReturnType(), t.Params()
		if stlbasic.Is[llvm.StructType](ret) || stlbasic.Is[llvm.Array](ret) {
			param = append([]llvm.Type{self.ctx.PointerType(ret)}, param...)
			ret = self.ctx.VoidType()
		}
		for i, p := range param {
			if stlbasic.Is[llvm.StructType](p) || stlbasic.Is[llvm.Array](p) {
				param[i] = self.ctx.PointerType(p)
			}
		}
		return self.module.NewFunction(name, self.ctx.FunctionType(false, ret, param...))
	case self.target.IsLinux():
		return self.module.NewFunction(name, self.ctx.FunctionType(false, t.ReturnType(), t.Params()...))
	default:
		panic("unreachable")
	}
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
