package codegen_ir

import (
	"bytes"
	"encoding/gob"
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	module2 "github.com/kkkunny/Sim/mir/pass/module"
	"github.com/kkkunny/Sim/runtime/types"
)

func (self *CodeGenerator) getExternFunction(name string, t mir.FuncType)*mir.Function{
	fn, ok := self.module.NamedFunction(name)
	if !ok {
		fn = self.module.NewFunction(name, t)
	}
	return fn
}

func (self *CodeGenerator) buildEqual(t hir.Type, l, r mir.Value, not bool) mir.Value {
	switch irType := hir.FlattenType(t).(type) {
	case *hir.SintType, *hir.UintType, *hir.FloatType, *hir.BoolType:
		return self.builder.BuildCmp(stlbasic.Ternary(!not, mir.CmpKindEQ, mir.CmpKindNE), l, r)
	case *hir.FuncType, *hir.RefType:
		return self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), l, r)
	case *hir.ArrayType:
		res := self.buildArrayEqual(irType, l, r)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	case *hir.TupleType, *hir.StructType:
		res := self.buildStructEqual(irType, l, r)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	case *hir.StringType:
		name := "sim_runtime_str_eq_str"
		ft := self.ctx.NewFuncType(false, self.ctx.Bool(), l.Type(), r.Type())
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
	case *hir.UnionType:
		res := self.buildUnionEqual(irType, l, r)
		if not {
			res = self.builder.BuildNot(res)
		}
		return res
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayEqual(irType *hir.ArrayType, l, r mir.Value) mir.Value {
	t := self.codegenTypeOnly(irType).(mir.ArrayType)
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
	cond = self.buildEqual(irType.Elem, self.buildArrayIndex(l, index, false), self.buildArrayIndex(r, index, false), false)
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

func (self *CodeGenerator) buildStructEqual(irType hir.Type, l, r mir.Value) mir.Value {
	_, isTuple := irType.(*hir.TupleType)
	t := stlbasic.TernaryAction(isTuple, func() mir.StructType {
		return self.codegenTypeOnly(irType).(mir.StructType)
	}, func() mir.StructType {
		return self.codegenTypeOnly(irType).(mir.StructType)
	})
	fields := stlbasic.TernaryAction(isTuple, func() dynarray.DynArray[hir.Type] {
		return dynarray.NewDynArrayWith(irType.(*hir.TupleType).Elems...)
	}, func() dynarray.DynArray[hir.Type] {
		values := irType.(*hir.StructType).Fields.Values()
		res := dynarray.NewDynArrayWithLength[hir.Type](values.Length())
		var i uint
		for iter:=values.Iterator(); iter.Next(); {
			res.Set(i, iter.Value().Type)
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

func (self *CodeGenerator) buildUnionEqual(irType *hir.UnionType, l, r mir.Value) mir.Value {
	key := fmt.Sprintf("equal:%s", irType.String())

	var f *mir.Function
	if !self.funcCache.ContainKey(key){
		curBlock := self.builder.Current()
		f = self.module.NewFunction("", self.ctx.NewFuncType(false, self.ctx.Bool(), l.Type(), r.Type()))
		lp, rp := f.Params()[0], f.Params()[1]

		self.builder.MoveTo(f.NewBlock())
		lk, rk := self.buildStructIndex(lp, 1, false), self.buildStructIndex(rp, 1, false)
		falseBlock, nextBlock := f.NewBlock(), f.NewBlock()
		self.builder.BuildCondJump(self.builder.BuildCmp(mir.CmpKindNE, lk, rk), falseBlock, nextBlock)

		self.builder.MoveTo(falseBlock)
		self.builder.BuildReturn(mir.Bool(self.ctx, false))

		self.builder.MoveTo(nextBlock)
		lvp, rvp := self.buildStructIndex(lp, 0, true), self.buildStructIndex(rp, 0, true)
		for i, elemIr := range irType.Elems{
			if i < len(irType.Elems) - 1 {
				var equalBlock *mir.Block
				equalBlock, nextBlock = f.NewBlock(), f.NewBlock()
				self.builder.BuildCondJump(self.builder.BuildCmp(mir.CmpKindEQ, lk, mir.NewInt(lk.Type().(mir.IntType), int64(i))), equalBlock, nextBlock)
				self.builder.MoveTo(equalBlock)
			}
			lv := self.builder.BuildLoad(self.builder.BuildPtrToPtr(lvp, self.ctx.NewPtrType(self.codegenTypeOnly(elemIr))))
			rv := self.builder.BuildLoad(self.builder.BuildPtrToPtr(rvp, self.ctx.NewPtrType(self.codegenTypeOnly(elemIr))))
			self.builder.BuildReturn(self.buildEqual(elemIr, lv, rv, false))
			if i < len(irType.Elems) - 1 {
				self.builder.MoveTo(nextBlock)
			}
		}
		self.builder.MoveTo(curBlock)
	}else{
		f = self.funcCache.Get(key)
	}

	return self.builder.BuildCall(f, l, r)
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
		mainFn = self.module.NewFunction("main", self.ctx.NewFuncType(false, self.ctx.U8()))
		mainFn.NewBlock()
	}
	return mainFn
}

func (self *CodeGenerator) getInitFunction() *mir.Function {
	initFn, ok := self.module.NamedFunction("sim_runtime_init")
	if !ok {
		initFn = self.module.NewFunction("sim_runtime_init", self.ctx.NewFuncType(false, self.ctx.Void()))
		initFn.SetAttribute(mir.FunctionAttributeInit)
		initFn.NewBlock()
	}
	return initFn
}

func (self *CodeGenerator) constString(s string) mir.Const {
	st, _ := self.codegenStringType()
	if !self.strings.ContainKey(s) {
		self.strings.Set(s, self.module.NewConstant("", mir.NewString(self.ctx, s)))
	}
	return mir.NewStruct(
		st,
		mir.NewArrayIndex(self.strings.Get(s), mir.NewInt(self.ctx.Usize(), 0)),
		mir.NewInt(self.ctx.Usize(), int64(len(s))),
	)
}

func (self *CodeGenerator) buildCovertUnionIndex(src, dst *types.UnionType, index mir.Value)mir.Value{
	strType, _ := self.codegenStringType()
	fn := self.getExternFunction("sim_runtime_covert_union_index", self.ctx.NewFuncType(false, self.ctx.U8(), strType, strType, self.ctx.U8()))

	gob.Register(new(types.EmptyType))
	gob.Register(new(types.BoolType))
	gob.Register(new(types.StringType))
	gob.Register(new(types.SintType))
	gob.Register(new(types.UintType))
	gob.Register(new(types.FloatType))
	gob.Register(new(types.RefType))
	gob.Register(new(types.FuncType))
	gob.Register(new(types.ArrayType))
	gob.Register(new(types.TupleType))
	gob.Register(new(types.UnionType))
	gob.Register(new(types.StructType))

	var srcStr, dstStr bytes.Buffer
	stlerror.Must(gob.NewEncoder(&srcStr).Encode(src))
	stlerror.Must(gob.NewEncoder(&dstStr).Encode(dst))
	return self.builder.BuildCall(fn, self.constString(srcStr.String()), self.constString(dstStr.String()), index)
}

func (self *CodeGenerator) buildCheckUnionType(src, dst *types.UnionType, index mir.Value)mir.Value{
	strType, _ := self.codegenStringType()
	fn := self.getExternFunction("sim_runtime_check_union_type", self.ctx.NewFuncType(false, self.ctx.Bool(), strType, strType, self.ctx.U8()))

	gob.Register(new(types.EmptyType))
	gob.Register(new(types.BoolType))
	gob.Register(new(types.StringType))
	gob.Register(new(types.SintType))
	gob.Register(new(types.UintType))
	gob.Register(new(types.FloatType))
	gob.Register(new(types.RefType))
	gob.Register(new(types.FuncType))
	gob.Register(new(types.ArrayType))
	gob.Register(new(types.TupleType))
	gob.Register(new(types.UnionType))
	gob.Register(new(types.StructType))

	var srcStr, dstStr bytes.Buffer
	stlerror.Must(gob.NewEncoder(&srcStr).Encode(src))
	stlerror.Must(gob.NewEncoder(&dstStr).Encode(dst))
	return self.builder.BuildCall(fn, self.constString(srcStr.String()), self.constString(dstStr.String()), index)
}

func (self *CodeGenerator) buildPanic(s string){
	strType, _ := self.codegenStringType()
	fn := self.getExternFunction("sim_runtime_panic", self.ctx.NewFuncType(false, self.ctx.Void(), strType))
	self.builder.BuildCall(fn, self.constString(s))
	self.builder.BuildUnreachable()
}

func (self *CodeGenerator) buildCheckZero(v mir.Value){
	cond := self.builder.BuildCmp(mir.CmpKindEQ, v, mir.NewZero(v.Type()))
	f := self.builder.Current().Belong()
	panicBlock, endBlock := f.NewBlock(), f.NewBlock()
	self.builder.BuildCondJump(cond, panicBlock, endBlock)

	self.builder.MoveTo(panicBlock)
	self.buildPanic("zero exception")

	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) buildCheckIndex(index mir.Value, rangev uint64){
	cond := self.builder.BuildCmp(mir.CmpKindGE, index, mir.NewInt(index.Type().(mir.IntType), int64(rangev)))
	f := self.builder.Current().Belong()
	panicBlock, endBlock := f.NewBlock(), f.NewBlock()
	self.builder.BuildCondJump(cond, panicBlock, endBlock)

	self.builder.MoveTo(panicBlock)
	self.buildPanic("index out of range")

	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) buildMalloc(t mir.Type)mir.Value{
	fn := self.getExternFunction("sim_runtime_malloc", self.ctx.NewFuncType(false, self.ctx.NewPtrType(self.ctx.I8()), self.ctx.Usize()))
	size := stlmath.RoundTo(t.Size(), stlos.Size(t.Align())*stlos.Byte)
	ptr := self.builder.BuildCall(fn, mir.NewUint(self.ctx.Usize(), uint64(size/stlos.Byte)))
	return self.builder.BuildPtrToPtr(ptr, self.ctx.NewPtrType(t))
}

// CodegenIr 中间代码生成
func CodegenIr(target mir.Target, path stlos.FilePath) (*mir.Module, stlerror.Error) {
	means, err := analyse.Analyse(path)
	if err != nil{
		return nil, err
	}
	module := New(target, means).Codegen()
	module2.Run(module, module2.DeadCodeElimination)
	return module, nil
}
