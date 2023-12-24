package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

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
	case hir.Covert:
		return self.codegenCovert(expr)
	case *hir.Array:
		return self.codegenArray(expr)
	case *hir.Index:
		return self.codegenIndex(expr, load)
	case *hir.Tuple:
		return self.codegenTuple(expr)
	case *hir.Extract:
		return self.codegenExtract(expr, load)
	case *hir.Default:
		return self.codegenZero(expr.GetType())
	case *hir.Struct:
		return self.codegenStruct(expr)
	case *hir.Field:
		return self.codegenField(expr, load)
	case *hir.String:
		return self.codegenString(expr)
	case *hir.Union:
		return self.codegenUnion(expr, load)
	case *hir.UnionTypeJudgment:
		return self.codegenUnionTypeJudgment(expr)
	case *hir.UnUnion:
		return self.codegenUnUnion(expr)
	case *hir.WrapWithNull:
		return self.codegenWrapWithNull(expr, load)
	case *hir.CheckNull:
		return self.codegenCheckNull(expr)
	case *hir.MethodDef:
		// TODO: 闭包
		panic("unreachable")
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
	if tupleNode, ok := fromIr.(*hir.Tuple); ok{
		for i, l := range toIrs {
			self.codegenAssign(&hir.Assign{
				Left:  l,
				Right: tupleNode.Elems[i],
			})
		}
	}else{
		var unTuple func(from mir.Value, toNodes []hir.Expr)
		unTuple = func(from mir.Value, toNodes []hir.Expr) {
			for i, toNode := range toNodes{
				if toNodes, ok := toNode.(*hir.Tuple); ok{
					index := self.buildStructIndex(from, uint64(i))
					unTuple(index, toNodes.Elems)
				}else{
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
		return self.builder.BuildDiv(left, right)
	case *hir.NumRemNum:
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
	case *hir.GetPtr:
		return self.codegenExpr(ir.GetValue(), false)
	case *hir.GetValue:
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
	case *hir.Param, *hir.VarDef:
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
	if method, ok := ir.Func.(*hir.Method); ok{
		f := self.values.Get(method.Define)
		selfParam := self.codegenExpr(method.Self, true)
		args := lo.Map(ir.Args, func(item hir.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, append([]mir.Value{selfParam}, args...)...)
	} else{
		f := self.codegenExpr(ir.Func, true)
		args := lo.Map(ir.Args, func(item hir.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(ir hir.Covert) mir.Value {
	from := self.codegenExpr(ir.GetFrom(), true)
	to := self.codegenType(ir.GetType())

	switch ir.(type) {
	case *hir.Num2Num:
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(ir *hir.Array) mir.Value {
	elems := lo.Map(ir.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackArray(self.codegenType(ir.GetType()).(mir.ArrayType), elems...)
}

func (self *CodeGenerator) codegenIndex(ir *hir.Index, load bool) mir.Value {
	// TODO: 运行时异常：超出索引下标
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildArrayIndex(from, self.codegenExpr(ir.Index, true))
	if load && (stlbasic.Is[*mir.ArrayIndex](ptr) && ptr.(*mir.ArrayIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTuple(ir *hir.Tuple) mir.Value {
	elems := lo.Map(ir.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenType(ir.GetType()).(mir.StructType), elems...)
}

func (self *CodeGenerator) codegenExtract(ir *hir.Extract, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenZero(ir hir.Type) mir.Value {
	switch t := ir.(type) {
	case *hir.SelfType:
		return self.codegenZero(t.Self)
	case *hir.AliasType:
		return self.codegenZero(t.Target)
	default:
		return mir.NewZero(self.codegenType(t))
	}
}

func (self *CodeGenerator) codegenStruct(ir *hir.Struct) mir.Value {
	fields := lo.Map(ir.Fields, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenType(ir.Type).(mir.StructType), fields...)
}

func (self *CodeGenerator) codegenField(ir *hir.Field, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenString(ir *hir.String) mir.Value {
	st := self.codegenStringType()
	if !self.strings.ContainKey(ir.Value) {
		self.strings.Set(ir.Value, self.module.NewConstant("", mir.NewString(self.ctx, ir.Value)))
	}
	return mir.NewStruct(
		st,
		mir.NewArrayIndex(self.strings.Get(ir.Value), mir.NewInt(self.ctx.Usize(), 0)),
		mir.NewInt(self.ctx.Usize(), int64(len(ir.Value))),
	)
}

func (self *CodeGenerator) codegenUnion(ir *hir.Union, load bool) mir.Value {
	ut := self.codegenType(ir.Type).(mir.StructType)
	value := self.codegenExpr(ir.Value, true)
	ptr := self.builder.BuildAllocFromStack(ut)
	dataPtr := self.buildStructIndex(ptr, 0, true)
	dataPtr = self.builder.BuildPtrToPtr(dataPtr, self.ctx.NewPtrType(value.Type()))
	self.builder.BuildStore(value, dataPtr)
	index := hir.AsUnionType(ir.Type).GetElemIndex(ir.Value.GetType())
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(index)),
		self.buildStructIndex(ptr, 1, true),
	)
	if load {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenUnionTypeJudgment(ir *hir.UnionTypeJudgment) mir.Value {
	ut := self.codegenType(ir.Value.GetType()).(mir.StructType)
	typeIndex := self.buildStructIndex(self.codegenExpr(ir.Value, false), 1, false)
	index := hir.AsUnionType(ir.Value.GetType()).GetElemIndex(ir.Type)
	return self.builder.BuildCmp(mir.CmpKindEQ, typeIndex, mir.NewInt(ut.Elems()[1].(mir.IntType), int64(index)))
}

func (self *CodeGenerator) codegenUnUnion(ir *hir.UnUnion) mir.Value {
	value := self.codegenExpr(ir.Value, false)
	elemPtr := self.buildStructIndex(value, 0, true)
	elemPtr = self.builder.BuildPtrToPtr(elemPtr, self.ctx.NewPtrType(self.codegenType(ir.GetType())))
	return self.builder.BuildLoad(elemPtr)
}

func (self *CodeGenerator) codegenWrapWithNull(ir *hir.WrapWithNull, load bool) mir.Value {
	return self.codegenExpr(ir.Value, load)
}

func (self *CodeGenerator) codegenCheckNull(ir *hir.CheckNull) mir.Value {
	name := "sim_runtime_check_null"
	ptrType := self.ctx.NewPtrType(self.ctx.U8())
	ft := self.ctx.NewFuncType(ptrType, ptrType)
	f, ok := self.module.NamedFunction(name)
	if !ok {
		f = self.module.NewFunction(name, ft)
	}

	ptr := self.codegenExpr(ir.Value, true)
	return self.builder.BuildCall(f, ptr)
}
