package codegen_ir

import (
	"fmt"
	"slices"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/util"
)

func (self *CodeGenerator) codegenExpr(ir hir.Expr, load bool) mir.Value {
	switch expr := ir.(type) {
	case *hir.Integer:
		return self.codegenInteger(expr)
	case *hir.Float:
		return self.codegenFloat(expr)
	case *hir.Boolean:
		return self.codegenBool(expr)
	case *hir.Assign:
		self.codegenAssign(expr)
		return nil
	case hir.Binary:
		return self.codegenBinary(expr)
	case hir.Unary:
		return self.codegenUnary(expr, load)
	case hir.Ident:
		return self.codegenIdent(expr, load)
	case *hir.Call:
		return self.codegenCall(expr)
	case hir.TypeCovert:
		return self.codegenCovert(expr, load)
	case *hir.Array:
		return self.codegenArray(expr)
	case *hir.Index:
		return self.codegenIndex(expr, load)
	case *hir.Tuple:
		return self.codegenTuple(expr)
	case *hir.Extract:
		return self.codegenExtract(expr, load)
	case *hir.Default:
		return self.codegenDefault(expr.GetType())
	case *hir.Struct:
		return self.codegenStruct(expr)
	case *hir.GetField:
		return self.codegenField(expr, load)
	case *hir.String:
		return self.codegenString(expr)
	case *hir.TypeJudgment:
		return self.codegenTypeJudgment(expr)
	case *hir.Lambda:
		return self.codegenLambda(expr)
	case *hir.Method:
		return self.codegenMethod(expr)
	case *hir.Enum:
		return self.codegenGetEnumField(expr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(ir *hir.Integer) mir.Int {
	return mir.NewInt(self.codegenType(ir.Type).(mir.IntType), ir.Value.Int64())
}

func (self *CodeGenerator) codegenFloat(ir *hir.Float) *mir.Float {
	v, _ := ir.Value.Float64()
	return mir.NewFloat(self.codegenType(ir.Type).(mir.FloatType), v)
}

func (self *CodeGenerator) codegenBool(ir *hir.Boolean) *mir.Uint {
	return mir.Bool(self.ctx, ir.Value)
}

func (self *CodeGenerator) codegenAssign(ir *hir.Assign) {
	if l, ok := ir.Left.(*hir.Tuple); ok {
		self.codegenUnTuple(ir.Right, l.Elems)
	} else {
		left, right := self.codegenExpr(ir.GetLeft(), false), self.codegenExpr(ir.GetRight(), true)
		self.builder.BuildStore(right, left)
	}
}

func (self *CodeGenerator) codegenUnTuple(fromIr hir.Expr, toIrs []hir.Expr) {
	if tupleNode, ok := fromIr.(*hir.Tuple); ok {
		for i, l := range toIrs {
			self.codegenAssign(&hir.Assign{
				Left:  l,
				Right: tupleNode.Elems[i],
			})
		}
	} else {
		var unTuple func(from mir.Value, toNodes []hir.Expr)
		unTuple = func(from mir.Value, toNodes []hir.Expr) {
			for i, toNode := range toNodes {
				if toNodes, ok := toNode.(*hir.Tuple); ok {
					index := self.buildStructIndex(from, uint64(i))
					unTuple(index, toNodes.Elems)
				} else {
					value := self.buildStructIndex(from, uint64(i), false)
					to := self.codegenExpr(toNode, false)
					self.builder.BuildStore(value, to)
				}
			}
		}
		from := self.codegenExpr(fromIr, false)
		unTuple(from, toIrs)
	}
}

func (self *CodeGenerator) codegenBinary(ir hir.Binary) mir.Value {
	left, right := self.codegenExpr(ir.GetLeft(), true), self.codegenExpr(ir.GetRight(), true)
	switch ir.(type) {
	case *hir.IntAndInt, *hir.BoolAndBool:
		return self.builder.BuildAnd(left, right)
	case *hir.IntOrInt, *hir.BoolOrBool:
		return self.builder.BuildOr(left, right)
	case *hir.IntXorInt:
		return self.builder.BuildXor(left, right)
	case *hir.IntShlInt:
		return self.builder.BuildShl(left, right)
	case *hir.IntShrInt:
		return self.builder.BuildShr(left, right)
	case *hir.NumAddNum:
		return self.builder.BuildAdd(left, right)
	case *hir.NumSubNum:
		return self.builder.BuildSub(left, right)
	case *hir.NumMulNum:
		return self.builder.BuildMul(left, right)
	case *hir.NumDivNum:
		self.buildCheckZero(right)
		return self.builder.BuildDiv(left, right)
	case *hir.NumRemNum:
		self.buildCheckZero(right)
		return self.builder.BuildRem(left, right)
	case *hir.NumLtNum:
		return self.builder.BuildCmp(mir.CmpKindLT, left, right)
	case *hir.NumGtNum:
		return self.builder.BuildCmp(mir.CmpKindGT, left, right)
	case *hir.NumLeNum:
		return self.builder.BuildCmp(mir.CmpKindLE, left, right)
	case *hir.NumGeNum:
		return self.builder.BuildCmp(mir.CmpKindGE, left, right)
	case *hir.Equal:
		return self.buildEqual(ir.GetLeft().GetType(), left, right, false)
	case *hir.NotEqual:
		return self.buildEqual(ir.GetLeft().GetType(), left, right, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(ir hir.Unary, load bool) mir.Value {
	switch ir.(type) {
	case *hir.NumNegate:
		return self.builder.BuildNeg(self.codegenExpr(ir.GetValue(), true))
	case *hir.IntBitNegate, *hir.BoolNegate:
		return self.builder.BuildNot(self.codegenExpr(ir.GetValue(), true))
	case *hir.GetRef:
		return self.codegenExpr(ir.GetValue(), false)
	case *hir.DeRef:
		ptr := self.codegenExpr(ir.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(ir hir.Ident, load bool) mir.Value {
	if !self.lambdaCaptureMap.Empty() {
		if mapVal := self.lambdaCaptureMap.Peek().Get(ir); mapVal != nil {
			if !load {
				return mapVal
			}
			return self.builder.BuildLoad(mapVal)
		}
	}

	switch identNode := ir.(type) {
	case *hir.FuncDef:
		return self.values.Get(identNode).(*mir.Function)
	case *hir.MethodDef:
		return self.values.Get(&identNode.FuncDef).(*mir.Function)
	case hir.Variable:
		p := self.values.Get(identNode)
		if !load {
			return p
		}
		return self.builder.BuildLoad(p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(ir *hir.Call) mir.Value {
	f := self.codegenExpr(ir.Func, true)
	args := stlslices.Map(ir.Args, func(_ int, e hir.Expr) mir.Value {
		return self.codegenExpr(e, true)
	})
	if hir.IsType[*hir.LambdaType](ir.Func.GetType()) {
		ctxPtr := self.buildStructIndex(f, 2, false)
		cf := self.builder.Current().Belong()
		f1block, f2block, endblock := cf.NewBlock(), cf.NewBlock(), cf.NewBlock()
		self.builder.BuildCondJump(self.builder.BuildPtrEqual(mir.PtrEqualKindEQ, ctxPtr, mir.NewZero(ctxPtr.Type())), f1block, f2block)

		self.builder.MoveTo(f1block)
		f1ret := self.builder.BuildCall(self.buildStructIndex(f, 0, false), args...)
		self.builder.BuildUnCondJump(endblock)

		self.builder.MoveTo(f2block)
		f2ret := self.builder.BuildCall(self.buildStructIndex(f, 1, false), append([]mir.Value{ctxPtr}, args...)...)
		self.builder.BuildUnCondJump(endblock)

		self.builder.MoveTo(endblock)
		if f1ret.Type().Equal(self.ctx.Void()) {
			return f1ret
		} else {
			return self.builder.BuildPhi(f1ret.Type(), pair.NewPair[*mir.Block, mir.Value](f1block, f1ret), pair.NewPair[*mir.Block, mir.Value](f2block, f2ret))
		}
	} else {
		return self.builder.BuildCall(f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(ir hir.TypeCovert, load bool) mir.Value {
	switch ir.(type) {
	case *hir.DoNothingCovert:
		return self.codegenExpr(ir.GetFrom(), load)
	case *hir.Num2Num:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenType(ir.GetType())
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	case *hir.NoReturn2Any:
		v := self.codegenExpr(ir.GetFrom(), false)
		self.builder.BuildUnreachable()
		if ir.GetType().EqualTo(hir.NoThing) {
			return v
		} else {
			return mir.NewZero(self.codegenType(ir.GetType()))
		}
	case *hir.Func2Lambda:
		t := self.codegenType(ir.GetType()).(mir.StructType)
		f := self.codegenExpr(ir.GetFrom(), true)
		return self.builder.BuildPackStruct(t, f, mir.NewZero(t.Elems()[1]), mir.NewZero(t.Elems()[2]))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(ir *hir.Array) mir.Value {
	elems := stlslices.Map(ir.Elems, func(_ int, item hir.Expr) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackArray(self.codegenType(ir.GetType()).(mir.ArrayType), elems...)
}

func (self *CodeGenerator) codegenIndex(ir *hir.Index, load bool) mir.Value {
	at := self.codegenType(ir.From.GetType()).(mir.ArrayType)
	index := self.codegenExpr(ir.Index, true)
	self.buildCheckIndex(index, uint64(at.Length()))
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildArrayIndex(from, index)
	if load && (stlbasic.Is[*mir.ArrayIndex](ptr) && ptr.(*mir.ArrayIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTuple(ir *hir.Tuple) mir.Value {
	elems := stlslices.Map(ir.Elems, func(_ int, item hir.Expr) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenType(ir.GetType()).(mir.StructType), elems...)
}

func (self *CodeGenerator) codegenExtract(ir *hir.Extract, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenDefault(ir hir.Type) mir.Value {
	switch tir := hir.ToRuntimeType(ir).(type) {
	case *hir.NoThingType, *hir.EnumType:
		panic("unreachable")
	case *hir.RefType:
		if tir.Elem.EqualTo(self.hir.BuildinTypes.Str) {
			return self.constStringPtr("")
		}
		panic("unreachable")
	case *hir.SintType, *hir.UintType, *hir.FloatType:
		return mir.NewZero(self.codegenType(ir))
	case *hir.ArrayType:
		at := self.codegenArrayType(tir)
		if tir.Size == 0 {
			return mir.NewZero(at)
		}

		key := fmt.Sprintf("default:%s", tir.String())
		var fn *mir.Function
		if !self.funcCache.ContainKey(key) {
			curBlock := self.builder.Current()
			ft := self.ctx.NewFuncType(false, at)
			fn = self.module.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveTo(fn.NewBlock())

			arrayPtr := self.builder.BuildAllocFromStack(at)
			indexPtr := self.builder.BuildAllocFromStack(self.usizeType())
			self.builder.BuildStore(mir.NewZero(indexPtr.ElemType()), indexPtr)
			condBlock := fn.NewBlock()
			self.builder.BuildUnCondJump(condBlock)

			self.builder.MoveTo(condBlock)
			index := self.builder.BuildLoad(indexPtr)
			cond := self.builder.BuildCmp(mir.CmpKindLT, index, mir.NewInt(self.usizeType(), int64(tir.Size)))
			loopBlock, endBlock := fn.NewBlock(), fn.NewBlock()
			self.builder.BuildCondJump(cond, loopBlock, endBlock)

			self.builder.MoveTo(loopBlock)
			elemPtr := self.buildArrayIndex(arrayPtr, index, true)
			self.builder.BuildStore(self.codegenDefault(tir.Elem), elemPtr)
			self.builder.BuildStore(self.builder.BuildAdd(index, mir.NewInt(self.usizeType(), 1)), indexPtr)
			self.builder.BuildUnCondJump(condBlock)

			self.builder.MoveTo(endBlock)
			self.builder.BuildReturn(self.builder.BuildLoad(arrayPtr))

			self.builder.MoveTo(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return self.builder.BuildCall(fn)
	case *hir.TupleType:
		elems := stlslices.Map(hir.AsType[*hir.TupleType](tir).Elems, func(_ int, e hir.Type) mir.Value {
			return self.codegenDefault(e)
		})
		return self.builder.BuildPackStruct(self.codegenTupleType(tir), elems...)
	case *hir.CustomType:
		if self.hir.BuildinTypes.Default.HasBeImpled(tir) {
			return self.codegenCall(&hir.Call{Func: hir.LoopFindMethodWithNoCheck(tir, util.None[hir.Expr](), self.hir.BuildinTypes.Default.FirstMethodName()).MustValue()})
		}
		return self.codegenDefault(tir.Target)
	case *hir.StructType:
		elems := stlslices.Map(hir.AsType[*hir.StructType](tir).Fields.Values().ToSlice(), func(_ int, e hir.Field) mir.Value {
			return self.codegenDefault(e.Type)
		})
		return self.builder.BuildPackStruct(self.codegenStructType(tir), elems...)
	case *hir.FuncType:
		ft := self.codegenFuncType(tir)
		key := fmt.Sprintf("default:%s", tir.String())
		var fn *mir.Function
		if !self.funcCache.ContainKey(key) {
			curBlock := self.builder.Current()
			fn = self.module.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveTo(fn.NewBlock())
			if ft.Ret().Equal(self.ctx.Void()) {
				self.builder.BuildReturn()
			} else {
				self.builder.BuildReturn(self.codegenDefault(hir.AsType[*hir.FuncType](tir).Ret))
			}
			self.builder.MoveTo(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return fn
	case *hir.LambdaType:
		t := self.codegenLambdaType(tir)
		fn := self.codegenDefault(tir.ToFuncType())
		return self.builder.BuildPackStruct(t, fn, mir.NewZero(t.Elems()[1]), mir.NewZero(t.Elems()[2]))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(ir *hir.Struct) mir.Value {
	fields := stlslices.Map(ir.Fields, func(_ int, item hir.Expr) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenType(ir.Type).(mir.StructType), fields...)
}

func (self *CodeGenerator) codegenField(ir *hir.GetField, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenString(ir *hir.String) mir.Value {
	return self.constStringPtr(ir.Value)
}

func (self *CodeGenerator) codegenTypeJudgment(ir *hir.TypeJudgment) mir.Value {
	return mir.Bool(self.ctx, ir.Value.GetType().EqualTo(ir.Type))
}

func (self *CodeGenerator) codegenLambda(ir *hir.Lambda) mir.Value {
	t := self.codegenType(ir.GetType()).(mir.StructType)
	isSimpleFunc := len(ir.Context) == 0
	ft := stlbasic.Ternary(isSimpleFunc, t.Elems()[0], t.Elems()[1]).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Ret.EqualTo(hir.NoReturn) {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}

	if isSimpleFunc {
		preBlock := self.builder.Current()
		self.builder.MoveTo(f.NewBlock())
		for i, pir := range ir.Params {
			self.values.Set(pir, f.Params()[i])
		}

		block, _ := self.codegenBlock(ir.Body, nil)
		self.builder.BuildUnCondJump(block)
		self.builder.MoveTo(preBlock)

		return self.builder.BuildPackStruct(t, f, mir.NewZero(t.Elems()[1]), mir.NewZero(t.Elems()[2]))
	} else {
		ctxType := self.ctx.NewStructType(stlslices.Map(ir.Context, func(_ int, e hir.Ident) mir.Type {
			return self.ctx.NewPtrType(self.codegenType(e.GetType()))
		})...)
		externalCtxPtr := self.buildMalloc(ctxType)
		for i, identIr := range ir.Context {
			self.builder.BuildStore(
				self.codegenIdent(identIr, false),
				self.buildStructIndex(externalCtxPtr, uint64(i), true),
			)
		}

		preBlock := self.builder.Current()
		self.builder.MoveTo(f.NewBlock())
		for i, pir := range ir.Params {
			self.values.Set(pir, f.Params()[i+1])
		}

		captureMap := hashmap.NewHashMapWithCapacity[hir.Ident, mir.Value](uint(len(ir.Context)))
		innerCtxPtr := self.builder.BuildPtrToPtr(self.builder.BuildLoad(f.Params()[0]), self.ctx.NewPtrType(ctxType))
		for i, identIr := range ir.Context {
			captureMap.Set(identIr, self.buildStructIndex(innerCtxPtr, uint64(i), false))
		}

		self.lambdaCaptureMap.Push(captureMap)
		defer func() {
			self.lambdaCaptureMap.Pop()
		}()

		block, _ := self.codegenBlock(ir.Body, nil)
		self.builder.BuildUnCondJump(block)
		self.builder.MoveTo(preBlock)

		return self.builder.BuildPackStruct(t, mir.NewZero(t.Elems()[0]), f, self.builder.BuildPtrToPtr(externalCtxPtr, t.Elems()[2].(mir.PtrType)))
	}
}

func (self *CodeGenerator) codegenMethod(ir *hir.Method) mir.Value {
	t := self.codegenType(ir.GetType()).(mir.StructType)
	ft := t.Elems()[1].(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.Ret.EqualTo(hir.NoReturn) {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}

	ctxType := self.ctx.NewStructType(self.codegenType(ir.Self.GetType()))
	externalCtxPtr := self.buildMalloc(ctxType)
	self.builder.BuildStore(
		self.codegenExpr(ir.Self, true),
		self.buildStructIndex(externalCtxPtr, 0, true),
	)

	preBlock := self.builder.Current()
	self.builder.MoveTo(f.NewBlock())
	method := self.codegenIdent(ir.Define, true)
	innerCtxPtr := self.builder.BuildPtrToPtr(self.builder.BuildLoad(f.Params()[0]), self.ctx.NewPtrType(ctxType))
	selfVal := self.buildStructIndex(innerCtxPtr, 0, false)
	args := []mir.Value{selfVal}
	args = append(args, stlslices.Map(f.Params()[1:], func(_ int, e *mir.Param) mir.Value {
		return self.builder.BuildLoad(e)
	})...)
	ret := self.builder.BuildCall(method, args...)
	if ret.Type().Equal(self.ctx.Void()) {
		self.builder.BuildReturn()
	} else {
		self.builder.BuildReturn(ret)
	}

	self.builder.MoveTo(preBlock)
	return self.builder.BuildPackStruct(t, mir.NewZero(t.Elems()[0]), f, self.builder.BuildPtrToPtr(externalCtxPtr, t.Elems()[2].(mir.PtrType)))
}

func (self *CodeGenerator) codegenGetEnumField(ir *hir.Enum) mir.Value {
	etIr := hir.AsType[*hir.EnumType](ir.GetType())
	index := slices.Index(etIr.Fields.Keys().ToSlice(), ir.Field)
	if etIr.IsSimple() {
		return mir.NewInt(self.codegenType(etIr).(mir.IntType), int64(index))
	}

	ut := self.codegenType(ir.GetType()).(mir.StructType)
	value := self.codegenTuple(&hir.Tuple{Elems: ir.Elems})
	ptr := self.builder.BuildAllocFromStack(ut)
	dataPtr := self.buildStructIndex(ptr, 0, true)
	dataPtr = self.builder.BuildPtrToPtr(dataPtr, self.ctx.NewPtrType(value.Type()))
	self.builder.BuildStore(value, dataPtr)
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(index)),
		self.buildStructIndex(ptr, 1, true),
	)
	return self.builder.BuildLoad(ptr)
}
