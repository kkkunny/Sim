package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/pass"
)

func (self *CodeGenerator) buildEqual(t mean.Type, l, r mir.Value, not bool) mir.Value {
	switch meanType := t.(type) {
	case mean.IntType, *mean.BoolType, *mean.FloatType:
		return self.builder.BuildCmp(stlbasic.Ternary(!not, mir.CmpKindEQ, mir.CmpKindNE), l, r)
	case *mean.FuncType, *mean.PtrType, *mean.RefType:
		return self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), l, r)
	case *mean.ArrayType:
		res := self.buildArrayEqual(meanType, l, r)
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
		var res mir.Value = self.builder.BuildCall(f, l, r)
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
	condBlock := self.builder.Current().Belong().NewBlock()
	self.builder.BuildUnCondJump(condBlock)

	// cond
	self.builder.MoveTo(condBlock)
	index := self.builder.BuildLoad(indexPtr)
	cond := self.builder.BuildCmp(mir.CmpKindLT, index, mir.NewUint(self.ctx.Usize(), uint64(t.Length())))
	bodyBlock, outBlock := self.builder.Current().Belong().NewBlock(), self.builder.Current().Belong().NewBlock()
	self.builder.BuildCondJump(cond, bodyBlock, outBlock)

	// body
	self.builder.MoveTo(bodyBlock)
	cond = self.buildEqual(meanType.Elem, self.buildArrayIndex(l, index, false), self.buildArrayIndex(r, index, false), false)
	bodyEndBlock := self.builder.Current()
	actionBlock := bodyEndBlock.Belong().NewBlock()
	self.builder.BuildCondJump(cond, actionBlock, outBlock)

	// action
	self.builder.MoveTo(actionBlock)
	self.builder.BuildStore(self.builder.BuildAdd(index, mir.NewUint(self.ctx.Usize(), 1)), indexPtr)
	self.builder.BuildUnCondJump(condBlock)

	// out
	self.builder.MoveTo(outBlock)
	return self.builder.BuildPhi(
		self.ctx.Bool(),
		pair.NewPair[*mir.Block, mir.Value](condBlock, mir.Bool(self.ctx, true)),
		pair.NewPair[*mir.Block, mir.Value](bodyEndBlock, mir.Bool(self.ctx, false)),
	)
}

func (self *CodeGenerator) buildStructEqual(meanType mean.Type, l, r mir.Value) mir.Value {
	_, isTuple := meanType.(*mean.TupleType)
	t := stlbasic.TernaryAction(isTuple, func() mir.StructType {
		return self.codegenTupleType(meanType.(*mean.TupleType))
	}, func() mir.StructType {
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

	if len(t.Elems()) == 0 {
		return mir.Bool(self.ctx, true)
	}

	beginBlock := self.builder.Current()
	nextBlocks := dynarray.NewDynArrayWithLength[pair.Pair[mir.Value, *mir.Block]](uint(len(t.Elems())))
	srcBlocks := dynarray.NewDynArrayWithLength[*mir.Block](uint(len(t.Elems())))
	for i := uint32(0); i < uint32(len(t.Elems())); i++ {
		cond := self.buildEqual(fields.Get(uint(i)), self.buildStructIndex(l, uint64(i), false), self.buildStructIndex(r, uint64(i), false), false)
		nextBlock := beginBlock.Belong().NewBlock()
		nextBlocks.Set(uint(i), pair.NewPair(cond, nextBlock))
		srcBlocks.Set(uint(i), self.builder.Current())
		self.builder.MoveTo(nextBlock)
	}
	self.builder.MoveTo(beginBlock)

	endBlock := nextBlocks.Back().Second
	for iter := nextBlocks.Iterator(); iter.Next(); {
		item := iter.Value()
		if iter.HasNext() {
			self.builder.BuildCondJump(item.First, item.Second, endBlock)
		} else {
			self.builder.BuildUnCondJump(endBlock)
		}
		self.builder.MoveTo(item.Second)
	}

	phi := self.builder.BuildPhi(self.ctx.Bool())
	for iter := srcBlocks.Iterator(); iter.Next(); {
		phi.AddFroms(pair.NewPair[*mir.Block, mir.Value](iter.Value(), stlbasic.Ternary[mir.Value](iter.HasNext(), mir.Bool(self.ctx, false), nextBlocks.Back().First)))
	}
	return phi
}

func (self *CodeGenerator) buildArrayIndex(array, i mir.Value, expectPtr ...bool)mir.Value{
	var expectType mir.PtrType
	if ft := array.Type(); stlbasic.Is[mir.PtrType](ft){
		expectType = self.ctx.NewPtrType(ft.(mir.PtrType).Elem().(mir.ArrayType).Elem())
	}else{
		expectType = self.ctx.NewPtrType(ft.(mir.ArrayType).Elem())
	}
	value := self.builder.BuildArrayIndex(array, i)
	if len(expectPtr) != 0 && expectPtr[0] && !value.Type().Equal(expectType){
		ptr := self.builder.BuildAllocFromStack(expectType.Elem())
		self.builder.BuildStore(value, ptr)
		return ptr
	}else if len(expectPtr) != 0 && !expectPtr[0] && value.Type().Equal(expectType){
		return self.builder.BuildLoad(value)
	}else {
		return value
	}
}

func (self *CodeGenerator) buildStructIndex(st mir.Value, i uint64, expectPtr ...bool)mir.Value{
	var expectType mir.PtrType
	if ft := st.Type(); stlbasic.Is[mir.PtrType](ft){
		expectType = self.ctx.NewPtrType(ft.(mir.PtrType).Elem().(mir.StructType).Elems()[i])
	}else{
		expectType = self.ctx.NewPtrType(ft.(mir.StructType).Elems()[i])
	}
	value := self.builder.BuildStructIndex(st, i)
	if len(expectPtr) != 0 && expectPtr[0] && !value.Type().Equal(expectType){
		ptr := self.builder.BuildAllocFromStack(expectType.Elem())
		self.builder.BuildStore(value, ptr)
		return ptr
	}else if len(expectPtr) != 0 && !expectPtr[0] && value.Type().Equal(expectType){
		return self.builder.BuildLoad(value)
	}else {
		return value
	}
}

func (self *CodeGenerator) getMainFunction() *mir.Function {
	mainFn, ok := self.module.NamedFunction("main")
	if !ok {
		mainFn = self.module.NewFunction("main", self.ctx.NewFuncType(self.ctx.U8()))
		mainFn.NewBlock()
	}
	return mainFn
}

func (self *CodeGenerator) getInitFunction() *mir.Function {
	initFn, ok := self.module.NamedFunction("sim_runtime_init")
	if !ok {
		initFn = self.module.NewFunction("sim_runtime_init", self.ctx.NewFuncType(self.ctx.Void()))
		initFn.NewBlock()
	}
	return initFn
}

// CodegenIr 中间代码生成
func CodegenIr(target mir.Target, path string) (*mir.Module, stlerror.Error) {
	means, err := analyse.Analyse(path)
	if err != nil{
		return nil, err
	}
	module := New(target, means).Codegen()
	pass.Run(module, pass.DeadCodeElimination)
	return module, nil
}
