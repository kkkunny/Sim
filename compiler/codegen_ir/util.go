package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"
	stlerror "github.com/kkkunny/stl/error"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/analyse"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/mir"
	module2 "github.com/kkkunny/Sim/mir/pass/module"
)

func (self *CodeGenerator) getExternFunction(name string, t mir.FuncType) *mir.Function {
	fn, ok := self.module.NamedFunction(name)
	if !ok {
		fn = self.module.NewFunction(name, t)
	}
	return fn
}

func (self *CodeGenerator) buildEqual(t hir.Type, l, r mir.Value, not bool) mir.Value {
	switch irType := hir.ToRuntimeType(t).(type) {
	case *hir.SintType, *hir.UintType, *hir.FloatType, *hir.EnumType:
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
	case *hir.CustomType:
		switch {
		case irType.EqualTo(self.hir.BuildinTypes.Str):
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
		default:
			return self.buildEqual(irType.Target, l, r, not)
		}
	case *hir.LambdaType:
		f1p := self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), self.buildStructIndex(l, 0, false), self.buildStructIndex(r, 0, false))
		f2p := self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), self.buildStructIndex(l, 1, false), self.buildStructIndex(r, 1, false))
		pp := self.builder.BuildPtrEqual(stlbasic.Ternary(!not, mir.PtrEqualKindEQ, mir.PtrEqualKindNE), self.buildStructIndex(l, 2, false), self.buildStructIndex(r, 2, false))
		return stlbasic.TernaryAction(!not, func() mir.Value {
			return self.builder.BuildAnd(self.builder.BuildAnd(f1p, f2p), pp)
		}, func() mir.Value {
			return self.builder.BuildOr(self.builder.BuildOr(f1p, f2p), pp)
		})
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildArrayEqual(irType *hir.ArrayType, l, r mir.Value) mir.Value {
	t := self.codegenType(irType).(mir.ArrayType)
	if t.Length() == 0 {
		return mir.Bool(self.ctx, true)
	}

	indexPtr := self.builder.BuildAllocFromStack(self.usizeType())
	self.builder.BuildStore(mir.NewZero(indexPtr.ElemType()), indexPtr)
	condBlock := self.builder.Current().Belong().NewBlock()
	self.builder.BuildUnCondJump(condBlock)

	// cond
	self.builder.MoveTo(condBlock)
	index := self.builder.BuildLoad(indexPtr)
	cond := self.builder.BuildCmp(mir.CmpKindLT, index, mir.NewUint(self.usizeType(), uint64(t.Length())))
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
	self.builder.BuildStore(self.builder.BuildAdd(index, mir.NewUint(self.usizeType(), 1)), indexPtr)
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
		return self.codegenType(irType).(mir.StructType)
	}, func() mir.StructType {
		return self.codegenType(irType).(mir.StructType)
	})
	fields := stlbasic.TernaryAction(isTuple, func() dynarray.DynArray[hir.Type] {
		return dynarray.NewDynArrayWith(irType.(*hir.TupleType).Elems...)
	}, func() dynarray.DynArray[hir.Type] {
		values := irType.(*hir.StructType).Fields.Values()
		res := dynarray.NewDynArrayWithLength[hir.Type](values.Length())
		var i uint
		for iter := values.Iterator(); iter.Next(); {
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

func (self *CodeGenerator) buildArrayIndex(array, i mir.Value, expectPtr ...bool) mir.Value {
	var expectType mir.PtrType
	if ft := array.Type(); stlbasic.Is[mir.PtrType](ft) {
		expectType = self.ctx.NewPtrType(ft.(mir.PtrType).Elem().(mir.ArrayType).Elem())
	} else {
		expectType = self.ctx.NewPtrType(ft.(mir.ArrayType).Elem())
	}
	value := self.builder.BuildArrayIndex(array, i)
	if len(expectPtr) != 0 && expectPtr[0] && !value.Type().Equal(expectType) {
		ptr := self.builder.BuildAllocFromStack(expectType.Elem())
		self.builder.BuildStore(value, ptr)
		return ptr
	} else if len(expectPtr) != 0 && !expectPtr[0] && value.Type().Equal(expectType) {
		return self.builder.BuildLoad(value)
	} else {
		return value
	}
}

func (self *CodeGenerator) buildStructIndex(st mir.Value, i uint64, expectPtr ...bool) mir.Value {
	var expectType mir.PtrType
	if ft := st.Type(); stlbasic.Is[mir.PtrType](ft) {
		expectType = self.ctx.NewPtrType(ft.(mir.PtrType).Elem().(mir.StructType).Elems()[i])
	} else {
		expectType = self.ctx.NewPtrType(ft.(mir.StructType).Elems()[i])
	}
	value := self.builder.BuildStructIndex(st, i)
	if len(expectPtr) != 0 && expectPtr[0] && !value.Type().Equal(expectType) {
		ptr := self.builder.BuildAllocFromStack(expectType.Elem())
		self.builder.BuildStore(value, ptr)
		return ptr
	} else if len(expectPtr) != 0 && !expectPtr[0] && value.Type().Equal(expectType) {
		return self.builder.BuildLoad(value)
	} else {
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

func (self *CodeGenerator) constStringPtr(s string) mir.Const {
	if !self.strings.ContainKey(s) {
		st := self.codegenType(self.hir.BuildinTypes.Str).(mir.StructType)
		dataPtr := stlbasic.TernaryAction(s == "", func() mir.Const {
			return mir.NewZero(st.Elems()[0])
		}, func() mir.Const {
			return mir.NewArrayIndex(
				self.module.NewConstant("", mir.NewString(self.ctx, s)),
				mir.NewInt(self.usizeType(), 0),
			)
		})
		self.strings.Set(s, self.module.NewConstant("", mir.NewStruct(st, dataPtr, mir.NewInt(self.usizeType(), int64(len(s))))))
	}
	return self.strings.Get(s)
}

func (self *CodeGenerator) buildPanic(s string) {
	strType := self.codegenType(self.hir.BuildinTypes.Str).(mir.StructType)
	fn := self.getExternFunction("sim_runtime_panic", self.ctx.NewFuncType(false, self.ctx.Void(), self.ctx.NewPtrType(strType)))
	self.builder.BuildCall(fn, self.constStringPtr(s))
	self.builder.BuildUnreachable()
}

func (self *CodeGenerator) buildCheckZero(v mir.Value) {
	cond := self.builder.BuildCmp(mir.CmpKindEQ, v, mir.NewZero(v.Type()))
	f := self.builder.Current().Belong()
	panicBlock, endBlock := f.NewBlock(), f.NewBlock()
	self.builder.BuildCondJump(cond, panicBlock, endBlock)

	self.builder.MoveTo(panicBlock)
	self.buildPanic("zero exception")

	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) buildCheckIndex(index mir.Value, rangev uint64) {
	cond := self.builder.BuildCmp(mir.CmpKindGE, index, mir.NewInt(index.Type().(mir.IntType), int64(rangev)))
	f := self.builder.Current().Belong()
	panicBlock, endBlock := f.NewBlock(), f.NewBlock()
	self.builder.BuildCondJump(cond, panicBlock, endBlock)

	self.builder.MoveTo(panicBlock)
	self.buildPanic("index out of range")

	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) buildMalloc(t mir.Type) mir.Value {
	fn := self.getExternFunction("sim_runtime_malloc", self.ctx.NewFuncType(false, self.ctx.NewPtrType(self.ctx.I8()), self.ctx.Usize()))
	size := stlmath.RoundTo(t.Size(), stlos.Size(t.Align())*stlos.Byte)
	ptr := self.builder.BuildCall(fn, mir.NewUint(self.ctx.Usize(), uint64(size/stlos.Byte)))
	return self.builder.BuildPtrToPtr(ptr, self.ctx.NewPtrType(t))
}

func (self *CodeGenerator) usizeType() mir.UintType {
	return self.codegenType(self.hir.BuildinTypes.Usize).(mir.UintType)
}

func (self *CodeGenerator) ptrType() mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(self.hir.BuildinTypes.U8))
}

func (self *CodeGenerator) boolType() mir.UintType {
	return self.codegenType(self.hir.BuildinTypes.Bool).(mir.UintType)
}

// CodegenIr 中间代码生成
func CodegenIr(target mir.Target, path stlos.FilePath) (*mir.Module, stlerror.Error) {
	means, err := analyse.Analyse(path)
	if err != nil {
		return nil, err
	}
	module := New(target, means).Codegen()
	module2.Run(module, module2.DeadCodeElimination)
	return module, nil
}
