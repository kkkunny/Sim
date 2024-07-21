package codegen_ir

import (
	"fmt"
	"slices"

	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/util"
)

func (self *CodeGenerator) codegenExpr(ir hir.Expr, load bool) llvm.Value {
	switch expr := ir.(type) {
	case *hir.Integer:
		return self.codegenInteger(expr)
	case *hir.Float:
		return self.codegenFloat(expr)
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

func (self *CodeGenerator) codegenInteger(ir *hir.Integer) llvm.ConstInteger {
	return self.ctx.ConstInteger(self.codegenType(ir.Type).(llvm.IntegerType), ir.Value.Int64())
}

func (self *CodeGenerator) codegenFloat(ir *hir.Float) llvm.ConstFloat {
	v, _ := ir.Value.Float64()
	return self.ctx.ConstFloat(self.codegenType(ir.Type).(llvm.FloatType), v)
}

func (self *CodeGenerator) codegenAssign(ir *hir.Assign) {
	if l, ok := ir.Left.(*hir.Tuple); ok {
		self.codegenUnTuple(ir.Right, l.Elems)
	} else {
		left, right := self.codegenExpr(ir.GetLeft(), false), self.codegenExpr(ir.GetRight(), true)
		self.builder.CreateStore(right, left)
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
		var unTuple func(t hir.Type, from llvm.Value, toNodes []hir.Expr)
		unTuple = func(t hir.Type, from llvm.Value, toNodes []hir.Expr) {
			st := self.codegenType(t).(llvm.StructType)
			for i, toNode := range toNodes {
				if toNodes, ok := toNode.(*hir.Tuple); ok {
					index := self.buildStructIndex(st, from, uint(i))
					unTuple(hir.AsType[*hir.TupleType](t).Elems[i], index, toNodes.Elems)
				} else {
					value := self.buildStructIndex(st, from, uint(i), false)
					to := self.codegenExpr(toNode, false)
					self.builder.CreateStore(value, to)
				}
			}
		}
		from := self.codegenExpr(fromIr, false)
		unTuple(fromIr.GetType(), from, toIrs)
	}
}

func (self *CodeGenerator) codegenBinary(ir hir.Binary) llvm.Value {
	left, right := self.codegenExpr(ir.GetLeft(), true), self.codegenExpr(ir.GetRight(), true)
	switch ir.(type) {
	case *hir.IntAndInt, *hir.BoolAndBool:
		return self.builder.CreateAnd("", left, right)
	case *hir.IntOrInt, *hir.BoolOrBool:
		return self.builder.CreateOr("", left, right)
	case *hir.IntXorInt:
		return self.builder.CreateXor("", left, right)
	case *hir.IntShlInt:
		return self.builder.CreateShl("", left, right)
	case *hir.IntShrInt:
		if t := ir.GetType(); hir.IsType[*hir.SintType](t) {
			return self.builder.CreateAShr("", left, right)
		} else {
			return self.builder.CreateLShr("", left, right)
		}
	case *hir.NumAddNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFAdd("", left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateSAdd("", left, right)
		} else {
			return self.builder.CreateUAdd("", left, right)
		}
	case *hir.NumSubNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFSub("", left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateSSub("", left, right)
		} else {
			return self.builder.CreateUSub("", left, right)
		}
	case *hir.NumMulNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFMul("", left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateSMul("", left, right)
		} else {
			return self.builder.CreateUMul("", left, right)
		}
	case *hir.NumDivNum:
		self.buildCheckZero(right)
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFDiv("", left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateSDiv("", left, right)
		} else {
			return self.builder.CreateUDiv("", left, right)
		}
	case *hir.NumRemNum:
		self.buildCheckZero(right)
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFRem("", left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateSRem("", left, right)
		} else {
			return self.builder.CreateURem("", left, right)
		}
	case *hir.NumLtNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFloatCmp("", llvm.FloatOLT, left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateIntCmp("", llvm.IntSLT, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntULT, left, right)
		}
	case *hir.NumGtNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFloatCmp("", llvm.FloatOGT, left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateIntCmp("", llvm.IntSGT, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntUGT, left, right)
		}
	case *hir.NumLeNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFloatCmp("", llvm.FloatOLE, left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateIntCmp("", llvm.IntSLE, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntULE, left, right)
		}
	case *hir.NumGeNum:
		if t := ir.GetType(); hir.IsType[*hir.FloatType](t) {
			return self.builder.CreateFloatCmp("", llvm.FloatOGE, left, right)
		} else if hir.IsType[*hir.SintType](t) {
			return self.builder.CreateIntCmp("", llvm.IntSGE, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntUGE, left, right)
		}
	case *hir.Equal:
		return self.builder.CreateZExt("", self.buildEqual(ir.GetLeft().GetType(), left, right, false), self.boolType())
	case *hir.NotEqual:
		return self.builder.CreateZExt("", self.buildEqual(ir.GetLeft().GetType(), left, right, true), self.boolType())
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(ir hir.Unary, load bool) llvm.Value {
	switch ir.(type) {
	case *hir.NumNegate:
		return self.codegenBinary(&hir.NumSubNum{
			Left:  &hir.Default{Type: ir.GetValue().GetType()},
			Right: ir.GetValue(),
		})
	case *hir.IntBitNegate, *hir.BoolNegate:
		return self.builder.CreateNot("", self.codegenExpr(ir.GetValue(), true))
	case *hir.GetRef:
		return self.codegenExpr(ir.GetValue(), false)
	case *hir.DeRef:
		ptr := self.codegenExpr(ir.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.CreateLoad("", self.codegenType(ir.GetType()), ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(ir hir.Ident, load bool) llvm.Value {
	if !self.lambdaCaptureMap.Empty() {
		if mapVal := self.lambdaCaptureMap.Peek().Get(ir); mapVal != nil {
			if !load {
				return mapVal
			}
			return self.builder.CreateLoad("", self.codegenType(ir.GetType()), mapVal)
		}
	}

	switch identNode := ir.(type) {
	case *hir.FuncDef:
		return self.values.Get(identNode)
	case *hir.MethodDef:
		return self.values.Get(&identNode.FuncDef)
	case hir.Variable:
		p := self.values.Get(identNode)
		if !load {
			return p
		}
		return self.builder.CreateLoad("", self.codegenType(ir.GetType()), p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(ir *hir.Call) llvm.Value {
	ft1, st, ft2 := self.codegenCallableType(hir.AsType[hir.CallableType](ir.Func.GetType()))
	f := self.codegenExpr(ir.Func, true)
	args := stlslices.Map(ir.Args, func(_ int, e hir.Expr) llvm.Value {
		return self.codegenExpr(e, true)
	})
	if hir.IsType[*hir.LambdaType](ir.Func.GetType()) {
		ctxPtr := self.buildStructIndex(st, f, 2, false)
		cf := self.builder.CurrentBlock().Belong()
		f1block, f2block, endblock := cf.NewBlock(""), cf.NewBlock(""), cf.NewBlock("")
		self.builder.CreateCondBr(self.builder.CreateIntCmp("", llvm.IntEQ, ctxPtr, self.ctx.ConstZero(ctxPtr.Type())), f1block, f2block)

		self.builder.MoveToAfter(f1block)
		f1ret := self.builder.CreateCall("", ft1, self.buildStructIndex(st, f, 0, false), args...)
		self.builder.CreateBr(endblock)

		self.builder.MoveToAfter(f2block)
		f2ret := self.builder.CreateCall("", ft2, self.buildStructIndex(st, f, 1, false), append([]llvm.Value{ctxPtr}, args...)...)
		self.builder.CreateBr(endblock)

		self.builder.MoveToAfter(endblock)
		if f1ret.Type().Equal(self.ctx.VoidType()) {
			return f1ret
		} else {
			return self.builder.CreatePHI(
				"",
				f1ret.Type(),
				struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: f1ret, Block: f1block},
				struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: f2ret, Block: f2block},
			)
		}
	} else {
		return self.builder.CreateCall("", ft1, f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(ir hir.TypeCovert, load bool) llvm.Value {
	switch ir.(type) {
	case *hir.DoNothingCovert, *hir.Enum2Number, *hir.Number2Enum:
		return self.codegenExpr(ir.GetFrom(), load)
	case *hir.Num2Num:
		ft := ir.GetFrom().GetType()
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenType(ir.GetType())
		switch {
		case hir.IsIntType(ft) && hir.IsIntType(ir.GetType()):
			ftSize, toSize := from.Type().(llvm.IntegerType).Bits(), to.(llvm.IntegerType).Bits()
			if ftSize < toSize {
				if hir.IsType[*hir.SintType](ft) {
					return self.builder.CreateSExt("", from, to.(llvm.IntegerType))
				} else {
					return self.builder.CreateZExt("", from, to.(llvm.IntegerType))
				}
			} else if ftSize > toSize {
				return self.builder.CreateTrunc("", from, to.(llvm.IntegerType))
			} else {
				return from
			}
		case hir.IsType[*hir.FloatType](ft) && hir.IsType[*hir.FloatType](ir.GetType()):
			ftSize, toSize := from.Type().(llvm.FloatType).Kind(), to.(llvm.FloatType).Kind()
			if ftSize < toSize {
				return self.builder.CreateFPExt("", from, to.(llvm.FloatType))
			} else if ftSize > toSize {
				return self.builder.CreateFPTrunc("", from, to.(llvm.FloatType))
			} else {
				return from
			}
		case hir.IsType[*hir.SintType](ft) && hir.IsType[*hir.FloatType](ir.GetType()):
			return self.builder.CreateSIToFP("", from, to.(llvm.FloatType))
		case hir.IsType[*hir.UintType](ft) && hir.IsType[*hir.FloatType](ir.GetType()):
			return self.builder.CreateUIToFP("", from, to.(llvm.FloatType))
		case hir.IsType[*hir.FloatType](ft) && hir.IsType[*hir.SintType](ir.GetType()):
			return self.builder.CreateFPToSI("", from, to.(llvm.IntegerType))
		case hir.IsType[*hir.FloatType](ft) && hir.IsType[*hir.UintType](ir.GetType()):
			return self.builder.CreateFPToUI("", from, to.(llvm.IntegerType))
		default:
			panic("unreachable")
		}
	case *hir.NoReturn2Any:
		v := self.codegenExpr(ir.GetFrom(), false)
		self.builder.CreateUnreachable()
		if ir.GetType().EqualTo(hir.NoThing) {
			return v
		} else {
			return self.ctx.ConstZero(self.codegenType(ir.GetType()))
		}
	case *hir.Func2Lambda:
		t := self.codegenType(ir.GetType()).(llvm.StructType)
		f := self.codegenExpr(ir.GetFrom(), true)
		return self.buildPackStruct(t, f, self.ctx.ConstZero(t.Elems()[1]), self.ctx.ConstZero(t.Elems()[2]))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(ir *hir.Array) llvm.Value {
	elems := stlslices.Map(ir.Elems, func(_ int, item hir.Expr) llvm.Value {
		return self.codegenExpr(item, true)
	})
	return self.buildPackArray(self.codegenType(ir.GetType()).(llvm.ArrayType), elems...)
}

func (self *CodeGenerator) codegenIndex(ir *hir.Index, load bool) llvm.Value {
	at := self.codegenType(ir.From.GetType()).(llvm.ArrayType)
	index := self.codegenExpr(ir.Index, true)
	self.buildCheckIndex(index, uint64(at.Capacity()))
	from := self.codegenExpr(ir.From, false)
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.buildArrayIndex(at, from, index, expectPtr...)
}

func (self *CodeGenerator) codegenTuple(ir *hir.Tuple) llvm.Value {
	elems := stlslices.Map(ir.Elems, func(_ int, item hir.Expr) llvm.Value {
		return self.codegenExpr(item, true)
	})
	return self.buildPackStruct(self.codegenType(ir.GetType()).(llvm.StructType), elems...)
}

func (self *CodeGenerator) codegenExtract(ir *hir.Extract, load bool) llvm.Value {
	from := self.codegenExpr(ir.From, false)
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.buildStructIndex(self.codegenType(ir.From.GetType()).(llvm.StructType), from, ir.Index, expectPtr...)
}

func (self *CodeGenerator) codegenDefault(ir hir.Type) llvm.Value {
	switch tir := hir.ToRuntimeType(ir).(type) {
	case *hir.NoThingType, *hir.EnumType:
		panic("unreachable")
	case *hir.RefType:
		if tir.Elem.EqualTo(self.hir.BuildinTypes.Str) {
			return self.constStringPtr("")
		}
		panic("unreachable")
	case *hir.SintType, *hir.UintType, *hir.FloatType:
		return self.ctx.ConstZero(self.codegenType(ir))
	case *hir.ArrayType:
		at := self.codegenArrayType(tir)
		if tir.Size == 0 {
			return self.ctx.ConstZero(at)
		}

		key := fmt.Sprintf("default:%s", tir.String())
		var fn llvm.Function
		if !self.funcCache.ContainKey(key) {
			curBlock := self.builder.CurrentBlock()
			ft := self.ctx.FunctionType(false, at)
			fn = self.module.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveToAfter(fn.NewBlock(""))

			arrayPtr := self.builder.CreateAlloca("", at)
			indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
			self.builder.CreateStore(self.ctx.ConstZero(self.ctx.IntPtrType(self.target)), indexPtr)
			condBlock := fn.NewBlock("")
			self.builder.CreateBr(condBlock)

			self.builder.MoveToAfter(condBlock)
			index := self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr)
			cond := self.builder.CreateIntCmp("", llvm.IntULT, index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(tir.Size)))
			loopBlock, endBlock := fn.NewBlock(""), fn.NewBlock("")
			self.builder.CreateCondBr(cond, loopBlock, endBlock)

			self.builder.MoveToAfter(loopBlock)
			elemPtr := self.buildArrayIndex(at, arrayPtr, index, true)
			self.builder.CreateStore(self.codegenDefault(tir.Elem), elemPtr)
			self.builder.CreateStore(self.builder.CreateUAdd("", index, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
			self.builder.CreateBr(condBlock)

			self.builder.MoveToAfter(endBlock)
			self.builder.CreateRet(stlbasic.Ptr[llvm.Value](self.builder.CreateLoad("", at, arrayPtr)))

			self.builder.MoveToAfter(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return self.builder.CreateCall("", fn.FunctionType(), fn)
	case *hir.TupleType:
		elems := stlslices.Map(hir.AsType[*hir.TupleType](tir).Elems, func(_ int, e hir.Type) llvm.Value {
			return self.codegenDefault(e)
		})
		return self.buildPackStruct(self.codegenTupleType(tir), elems...)
	case *hir.CustomType:
		if self.hir.BuildinTypes.Default.HasBeImpled(tir) {
			return self.codegenCall(&hir.Call{Func: hir.LoopFindMethodWithNoCheck(tir, util.None[hir.Expr](), self.hir.BuildinTypes.Default.FirstMethodName()).MustValue()})
		}
		return self.codegenDefault(tir.Target)
	case *hir.StructType:
		elems := stlslices.Map(hir.AsType[*hir.StructType](tir).Fields.Values().ToSlice(), func(_ int, e hir.Field) llvm.Value {
			return self.codegenDefault(e.Type)
		})
		return self.buildPackStruct(self.codegenStructType(tir), elems...)
	case *hir.FuncType:
		ft := self.codegenFuncType(tir)
		key := fmt.Sprintf("default:%s", tir.String())
		var fn llvm.Function
		if !self.funcCache.ContainKey(key) {
			curBlock := self.builder.CurrentBlock()
			fn = self.module.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveToAfter(fn.NewBlock(""))
			if ft.ReturnType().Equal(self.ctx.VoidType()) {
				self.builder.CreateRet(nil)
			} else {
				self.builder.CreateRet(stlbasic.Ptr(self.codegenDefault(hir.AsType[*hir.FuncType](tir).Ret)))
			}
			self.builder.MoveToAfter(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return fn
	case *hir.LambdaType:
		t := self.codegenLambdaType(tir)
		fn := self.codegenDefault(tir.ToFuncType())
		return self.buildPackStruct(t, fn, self.ctx.ConstZero(t.Elems()[1]), self.ctx.ConstZero(t.Elems()[2]))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(ir *hir.Struct) llvm.Value {
	fields := stlslices.Map(ir.Fields, func(_ int, item hir.Expr) llvm.Value {
		return self.codegenExpr(item, true)
	})
	return self.buildPackStruct(self.codegenType(ir.Type).(llvm.StructType), fields...)
}

func (self *CodeGenerator) codegenField(ir *hir.GetField, load bool) llvm.Value {
	from := self.codegenExpr(ir.From, false)
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.buildStructIndex(self.codegenType(ir.From.GetType()).(llvm.StructType), from, ir.Index, expectPtr...)
}

func (self *CodeGenerator) codegenString(ir *hir.String) llvm.Value {
	return self.constStringPtr(ir.Value)
}

func (self *CodeGenerator) codegenTypeJudgment(ir *hir.TypeJudgment) llvm.Value {
	return self.builder.CreateZExt("", self.ctx.ConstBoolean(ir.Value.GetType().EqualTo(ir.Type)), self.boolType())
}

func (self *CodeGenerator) codegenLambda(ir *hir.Lambda) llvm.Value {
	ft1, st, ft2 := self.codegenCallableType(hir.AsType[hir.CallableType](ir.GetType()))
	isSimpleFunc := len(ir.Context) == 0
	ft := stlbasic.Ternary(isSimpleFunc, ft1, ft2)
	f := self.module.NewFunction("", ft)
	if ir.Ret.EqualTo(hir.NoReturn) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}

	if isSimpleFunc {
		preBlock := self.builder.CurrentBlock()
		self.builder.MoveToAfter(f.NewBlock(""))
		for i, pir := range ir.Params {
			p := self.builder.CreateAlloca("", self.codegenType(pir.Type))
			self.builder.CreateStore(f.GetParam(uint(i)), p)
			self.values.Set(pir, p)
		}

		block, _ := self.codegenBlock(ir.Body, nil)
		self.builder.CreateBr(block)
		self.builder.MoveToAfter(preBlock)

		return self.buildPackStruct(st, f, self.ctx.ConstZero(st.GetElem(1)), self.ctx.ConstZero(st.GetElem(2)))
	} else {
		ctxType := self.ctx.StructType(false, stlslices.Map(ir.Context, func(_ int, e hir.Ident) llvm.Type {
			return self.ctx.PointerType(self.codegenType(e.GetType()))
		})...)
		externalCtxPtr := self.buildMalloc(ctxType)
		for i, identIr := range ir.Context {
			self.builder.CreateStore(
				self.codegenIdent(identIr, false),
				self.buildStructIndex(ctxType, externalCtxPtr, uint(i), true),
			)
		}

		preBlock := self.builder.CurrentBlock()
		self.builder.MoveToAfter(f.NewBlock(""))
		for i, pir := range ir.Params {
			p := self.builder.CreateAlloca("", self.codegenType(pir.Type))
			self.builder.CreateStore(f.GetParam(uint(i+1)), p)
			self.values.Set(pir, p)
		}

		captureMap := hashmap.NewHashMapWithCapacity[hir.Ident, llvm.Value](uint(len(ir.Context)))
		innerCtxPtr := self.builder.CreateBitCast("", f.GetParam(0), self.ctx.PointerType(ctxType))
		for i, identIr := range ir.Context {
			captureMap.Set(identIr, self.buildStructIndex(ctxType, innerCtxPtr, uint(i), false))
		}

		self.lambdaCaptureMap.Push(captureMap)
		defer func() {
			self.lambdaCaptureMap.Pop()
		}()

		block, _ := self.codegenBlock(ir.Body, nil)
		self.builder.CreateBr(block)
		self.builder.MoveToAfter(preBlock)

		return self.buildPackStruct(st, self.ctx.ConstZero(st.GetElem(0)), f, self.builder.CreateBitCast("", externalCtxPtr, st.GetElem(2)))
	}
}

func (self *CodeGenerator) codegenMethod(ir *hir.Method) llvm.Value {
	_, st, ft2 := self.codegenCallableType(hir.AsType[hir.CallableType](ir.GetType()))
	f := self.module.NewFunction("", ft2)
	if ir.Define.Ret.EqualTo(hir.NoReturn) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}

	ctxType := self.ctx.StructType(false, self.codegenType(ir.Self.GetType()))
	externalCtxPtr := self.buildMalloc(ctxType)
	self.builder.CreateStore(
		self.codegenExpr(ir.Self, true),
		self.buildStructIndex(ctxType, externalCtxPtr, 0, true),
	)

	preBlock := self.builder.CurrentBlock()
	self.builder.MoveToAfter(f.NewBlock(""))
	method := self.codegenIdent(ir.Define, true)
	innerCtxPtr := self.builder.CreateBitCast("", f.GetParam(0), self.ctx.PointerType(ctxType))
	selfVal := self.buildStructIndex(ctxType, innerCtxPtr, 0, false)
	args := []llvm.Value{selfVal}
	args = append(args, stlslices.Map(f.Params()[1:], func(i int, e llvm.Param) llvm.Value {
		return e
	})...)
	ret := self.builder.CreateCall("", self.codegenFuncType(ir.Define.GetMethodType()), method, args...)
	if ret.Type().Equal(self.ctx.VoidType()) {
		self.builder.CreateRet(nil)
	} else {
		self.builder.CreateRet(stlbasic.Ptr[llvm.Value](ret))
	}

	self.builder.MoveToAfter(preBlock)
	return self.buildPackStruct(st, self.ctx.ConstZero(st.GetElem(0)), f, self.builder.CreateBitCast("", externalCtxPtr, st.GetElem(2)))
}

func (self *CodeGenerator) codegenGetEnumField(ir *hir.Enum) llvm.Value {
	etIr := hir.AsType[*hir.EnumType](ir.GetType())
	index := slices.Index(etIr.Fields.Keys().ToSlice(), ir.Field)
	if etIr.IsSimple() {
		return self.ctx.ConstInteger(self.codegenType(etIr).(llvm.IntegerType), int64(index))
	}

	ut := self.codegenType(ir.GetType()).(llvm.StructType)
	value := self.codegenTuple(&hir.Tuple{Elems: ir.Elems})
	ptr := self.builder.CreateAlloca("", ut)
	dataPtr := self.buildStructIndex(ut, ptr, 0, true)
	dataPtr = self.builder.CreateBitCast("", dataPtr, self.ctx.PointerType(value.Type()))
	self.builder.CreateStore(value, dataPtr)
	self.builder.CreateStore(
		self.ctx.ConstInteger(ut.GetElem(1).(llvm.IntegerType), int64(index)),
		self.buildStructIndex(ut, ptr, 1, true),
	)
	return self.builder.CreateLoad("", ut, ptr)
}
