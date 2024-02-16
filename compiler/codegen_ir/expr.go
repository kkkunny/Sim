package codegen_ir

import (
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
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
	case *hir.Union:
		return self.codegenUnion(expr, load)
	case *hir.TypeJudgment:
		return self.codegenTypeJudgment(expr)
	case *hir.UnUnion:
		return self.codegenUnUnion(expr)
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
	switch identNode := ir.(type) {
	case *hir.FuncDef:
		return self.values.Get(identNode).(*mir.Function)
	case hir.Variable:
		p := self.values.Get(identNode)
		if !load {
			return p
		}
		return self.builder.BuildLoad(p)
	case *hir.Method:
		// TODO: 闭包
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(ir *hir.Call) mir.Value {
	args := stlslices.Map(ir.Args, func(_ int, e hir.Expr) mir.Value {
		return self.codegenExpr(e, true)
	})
	var f, selfValue mir.Value
	switch fnIr := ir.Func.(type) {
	case *hir.Method:
		if !fnIr.Define.IsStatic() {
			selfValue = self.codegenExpr(stlbasic.IgnoreWith(fnIr.Self.Left()), !hir.IsType[*hir.RefType](fnIr.Define.Params[0].GetType()))
		}
		f = self.values.Get(&fnIr.Define.FuncDef)
	default:
		f = self.codegenExpr(ir.Func, true)
	}
	if selfValue != nil {
		var selfRef mir.Value
		if expectSelfRefType := f.Type().(mir.FuncType).Params()[0]; !selfValue.Type().Equal(expectSelfRefType) {
			selfRef = self.builder.BuildAllocFromStack(selfValue.Type())
			self.builder.BuildStore(selfValue, selfRef)
		} else {
			selfRef = selfValue
		}
		args = append([]mir.Value{selfRef}, args...)
	}
	return self.builder.BuildCall(f, args...)
}

func (self *CodeGenerator) codegenCovert(ir hir.TypeCovert, load bool) mir.Value {
	switch ir.(type) {
	case *hir.DoNothingCovert:
		return self.codegenExpr(ir.GetFrom(), load)
	case *hir.Num2Num:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenType(ir.GetType())
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	case *hir.ShrinkUnion:
		from := self.codegenExpr(ir.GetFrom(), false)
		srcData := self.buildStructIndex(from, 0, false)
		srcIndex := self.buildStructIndex(from, 1, false)
		newIndex := self.buildCovertUnionIndex(hir.AsType[*hir.UnionType](ir.GetFrom().GetType()), hir.AsType[*hir.UnionType](ir.GetType()), srcIndex)
		ptr := self.builder.BuildAllocFromStack(self.codegenType(ir.GetType()))
		newDataPtr := self.builder.BuildPtrToPtr(self.buildStructIndex(ptr, 0, true), self.ctx.NewPtrType(srcData.Type()))
		self.builder.BuildStore(srcData, newDataPtr)
		newIndexPtr := self.buildStructIndex(ptr, 1, true)
		self.builder.BuildStore(newIndex, newIndexPtr)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	case *hir.ExpandUnion:
		from := self.codegenExpr(ir.GetFrom(), false)
		srcData := self.buildStructIndex(from, 0, false)
		srcIndex := self.buildStructIndex(from, 1, false)
		newIndex := self.buildCovertUnionIndex(hir.AsType[*hir.UnionType](ir.GetFrom().GetType()), hir.AsType[*hir.UnionType](ir.GetType()), srcIndex)
		ptr := self.builder.BuildAllocFromStack(self.codegenType(ir.GetType()))
		newDataPtr := self.builder.BuildPtrToPtr(self.buildStructIndex(ptr, 0, true), self.ctx.NewPtrType(srcData.Type()))
		self.builder.BuildStore(srcData, newDataPtr)
		newIndexPtr := self.buildStructIndex(ptr, 1, true)
		self.builder.BuildStore(newIndex, newIndexPtr)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	case *hir.NoReturn2Any:
		self.codegenExpr(ir.GetFrom(), false)
		self.builder.BuildUnreachable()
		return mir.NewZero(self.codegenType(ir.GetType()))
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
	case *hir.NoThingType:
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
			indexPtr := self.builder.BuildAllocFromStack(self.codegenUsizeType())
			self.builder.BuildStore(mir.NewZero(indexPtr.ElemType()), indexPtr)
			condBlock := fn.NewBlock()
			self.builder.BuildUnCondJump(condBlock)

			self.builder.MoveTo(condBlock)
			index := self.builder.BuildLoad(indexPtr)
			cond := self.builder.BuildCmp(mir.CmpKindLT, index, mir.NewInt(self.codegenUsizeType(), int64(tir.Size)))
			loopBlock, endBlock := fn.NewBlock(), fn.NewBlock()
			self.builder.BuildCondJump(cond, loopBlock, endBlock)

			self.builder.MoveTo(loopBlock)
			elemPtr := self.buildArrayIndex(arrayPtr, index, true)
			self.builder.BuildStore(self.codegenDefault(tir.Elem), elemPtr)
			self.builder.BuildStore(self.builder.BuildAdd(index, mir.NewInt(self.codegenUsizeType(), 1)), indexPtr)
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
			return self.codegenCall(&hir.Call{Func: &hir.Method{
				Self:   either.Right[hir.Expr, *hir.CustomType](tir),
				Define: tir.Methods.Get(self.hir.BuildinTypes.Default.Methods.Values().Front().Name).(*hir.MethodDef),
			}})
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
	case *hir.UnionType:
		ut := self.codegenUnionType(tir)
		val := self.codegenDefault(hir.AsType[*hir.UnionType](tir).Elems[0])
		index := mir.NewInt(ut.Elems()[1].(mir.IntType), 0)
		return self.builder.BuildPackStruct(ut, val, index)
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

func (self *CodeGenerator) codegenUnion(ir *hir.Union, load bool) mir.Value {
	ut := self.codegenType(ir.Type).(mir.StructType)

	value := self.codegenExpr(ir.Value, true)
	ptr := self.builder.BuildAllocFromStack(ut)
	dataPtr := self.buildStructIndex(ptr, 0, true)
	dataPtr = self.builder.BuildPtrToPtr(dataPtr, self.ctx.NewPtrType(value.Type()))
	self.builder.BuildStore(value, dataPtr)

	index := hir.AsType[*hir.UnionType](ir.Type).IndexElem(ir.Value.GetType())
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(index)),
		self.buildStructIndex(ptr, 1, true),
	)

	if load {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTypeJudgment(ir *hir.TypeJudgment) mir.Value {
	srcTObj := self.codegenType(ir.Value.GetType())
	switch {
	case hir.IsType[*hir.UnionType](ir.Value.GetType()) && hir.AsType[*hir.UnionType](ir.Value.GetType()).Contain(ir.Type):
		srcT := srcTObj.(mir.StructType)

		from := self.codegenExpr(ir.Value, false)
		srcIndex := self.buildStructIndex(from, 1, false)

		if dstUt, ok := hir.TryType[*hir.UnionType](ir.Type); ok {
			return self.buildCheckUnionType(hir.AsType[*hir.UnionType](ir.Value.GetType()), dstUt, srcIndex)
		} else {
			targetIndex := hir.AsType[*hir.UnionType](ir.Value.GetType()).IndexElem(ir.Type)
			return self.builder.BuildCmp(mir.CmpKindEQ, srcIndex, mir.NewInt(srcT.Elems()[1].(mir.IntType), int64(targetIndex)))
		}
	default:
		return mir.Bool(self.ctx, ir.Value.GetType().EqualTo(ir.Type))
	}
}

func (self *CodeGenerator) codegenUnUnion(ir *hir.UnUnion) mir.Value {
	value := self.codegenExpr(ir.Value, false)
	elemPtr := self.buildStructIndex(value, 0, true)
	elemPtr = self.builder.BuildPtrToPtr(elemPtr, self.ctx.NewPtrType(self.codegenType(ir.GetType())))
	return self.builder.BuildLoad(elemPtr)
}
