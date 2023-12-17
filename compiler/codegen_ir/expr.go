package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenExpr(node hir.Expr, load bool) mir.Value {
	switch exprNode := node.(type) {
	case *hir.Integer:
		return self.codegenInteger(exprNode)
	case *hir.Float:
		return self.codegenFloat(exprNode)
	case *hir.Boolean:
		return self.codegenBool(exprNode)
	case *hir.Assign:
		self.codegenAssign(exprNode)
		return nil
	case hir.Binary:
		return self.codegenBinary(exprNode)
	case hir.Unary:
		return self.codegenUnary(exprNode, load)
	case hir.Ident:
		return self.codegenIdent(exprNode, load)
	case *hir.Call:
		return self.codegenCall(exprNode)
	case hir.Covert:
		return self.codegenCovert(exprNode)
	case *hir.Array:
		return self.codegenArray(exprNode)
	case *hir.Index:
		return self.codegenIndex(exprNode, load)
	case *hir.Tuple:
		return self.codegenTuple(exprNode)
	case *hir.Extract:
		return self.codegenExtract(exprNode, load)
	case *hir.Default:
		return self.codegenZero(exprNode.GetType())
	case *hir.Struct:
		return self.codegenStruct(exprNode)
	case *hir.Field:
		return self.codegenField(exprNode, load)
	case *hir.String:
		return self.codegenString(exprNode)
	case *hir.Union:
		return self.codegenUnion(exprNode, load)
	case *hir.UnionTypeJudgment:
		return self.codegenUnionTypeJudgment(exprNode)
	case *hir.UnUnion:
		return self.codegenUnUnion(exprNode)
	case *hir.WrapWithNull:
		return self.codegenWrapWithNull(exprNode, load)
	case *hir.CheckNull:
		return self.codegenCheckNull(exprNode)
	case *hir.MethodDef:
		// TODO: 闭包
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *hir.Integer) mir.Int {
	return mir.NewInt(self.codegenIntType(node.Type), node.Value.Int64())
}

func (self *CodeGenerator) codegenFloat(node *hir.Float) *mir.Float {
	v, _ := node.Value.Float64()
	return mir.NewFloat(self.codegenFloatType(node.Type), v)
}

func (self *CodeGenerator) codegenBool(node *hir.Boolean) *mir.Uint {
	return mir.Bool(self.ctx, node.Value)
}

func (self *CodeGenerator) codegenAssign(node *hir.Assign) {
	if l, ok := node.Left.(*hir.Tuple); ok {
		self.codegenUnTuple(node.Right, l.Elems)
	} else {
		left, right := self.codegenExpr(node.GetLeft(), false), self.codegenExpr(node.GetRight(), true)
		self.builder.BuildStore(right, left)
	}
}

func (self *CodeGenerator) codegenUnTuple(fromNode hir.Expr, toNodes []hir.Expr) {
	if tupleNode, ok := fromNode.(*hir.Tuple); ok{
		for i, l := range toNodes {
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
		from := self.codegenExpr(fromNode, false)
		unTuple(from, toNodes)
	}
}

func (self *CodeGenerator) codegenBinary(node hir.Binary) mir.Value {
	left, right := self.codegenExpr(node.GetLeft(), true), self.codegenExpr(node.GetRight(), true)
	switch node.(type) {
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
		return self.buildEqual(node.GetLeft().GetType(), left, right, false)
	case *hir.NotEqual:
		return self.buildEqual(node.GetLeft().GetType(), left, right, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(node hir.Unary, load bool) mir.Value {
	switch node.(type) {
	case *hir.NumNegate:
		return self.builder.BuildNeg(self.codegenExpr(node.GetValue(), true))
	case *hir.IntBitNegate, *hir.BoolNegate:
		return self.builder.BuildNot(self.codegenExpr(node.GetValue(), true))
	case *hir.GetPtr:
		return self.codegenExpr(node.GetValue(), false)
	case *hir.GetValue:
		ptr := self.codegenExpr(node.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(node hir.Ident, load bool) mir.Value {
	switch identNode := node.(type) {
	case *hir.FuncDef,*hir.GenericFuncInstance:
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

func (self *CodeGenerator) codegenCall(node *hir.Call) mir.Value {
	if method, ok := node.Func.(*hir.Method); ok{
		f := self.values.Get(method.Define)
		selfParam := self.codegenExpr(method.Self, true)
		args := lo.Map(node.Args, func(item hir.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, append([]mir.Value{selfParam}, args...)...)
	} else{
		f := self.codegenExpr(node.Func, true)
		args := lo.Map(node.Args, func(item hir.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(node hir.Covert) mir.Value {
	from := self.codegenExpr(node.GetFrom(), true)
	to := self.codegenType(node.GetType())

	switch node.(type) {
	case *hir.Num2Num:
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(node *hir.Array) mir.Value {
	elems := lo.Map(node.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackArray(self.codegenArrayType(node.Type), elems...)
}

func (self *CodeGenerator) codegenIndex(node *hir.Index, load bool) mir.Value {
	// TODO: 运行时异常：超出索引下标
	from := self.codegenExpr(node.From, false)
	ptr := self.buildArrayIndex(from, self.codegenExpr(node.Index, true))
	if load && (stlbasic.Is[*mir.ArrayIndex](ptr) && ptr.(*mir.ArrayIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTuple(node *hir.Tuple) mir.Value {
	elems := lo.Map(node.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenTupleType(node.GetType().(*hir.TupleType)), elems...)
}

func (self *CodeGenerator) codegenExtract(node *hir.Extract, load bool) mir.Value {
	from := self.codegenExpr(node.From, false)
	ptr := self.buildStructIndex(from, uint64(node.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenZero(tNode hir.Type) mir.Value {
	switch ttNode := tNode.(type) {
	case hir.NumberType, *hir.BoolType, *hir.StringType, *hir.PtrType, *hir.RefType:
		return mir.NewZero(self.codegenType(ttNode))
	case *hir.ArrayType, *hir.StructType, *hir.UnionType:
		// TODO: 复杂类型default值
		return mir.NewZero(self.codegenType(ttNode))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(node *hir.Struct) mir.Value {
	fields := lo.Map(node.Fields, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenStructType(node.Type), fields...)
}

func (self *CodeGenerator) codegenField(node *hir.Field, load bool) mir.Value {
	from := self.codegenExpr(node.From, false)
	ptr := self.buildStructIndex(from, uint64(node.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()){
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenString(node *hir.String) mir.Value {
	st := self.codegenStringType()
	if !self.strings.ContainKey(node.Value) {
		self.strings.Set(node.Value, self.module.NewConstant("", mir.NewString(self.ctx, node.Value)))
	}
	return mir.NewStruct(
		st,
		mir.NewArrayIndex(self.strings.Get(node.Value), mir.NewInt(self.ctx.Usize(), 0)),
		mir.NewInt(self.ctx.Usize(), int64(len(node.Value))),
	)
}

func (self *CodeGenerator) codegenUnion(node *hir.Union, load bool) mir.Value {
	ut := self.codegenUnionType(node.Type)
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.BuildAllocFromStack(ut)
	dataPtr := self.buildStructIndex(ptr, 0, true)
	dataPtr = self.builder.BuildPtrToPtr(dataPtr, self.ctx.NewPtrType(value.Type()))
	self.builder.BuildStore(value, dataPtr)
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(node.Type.GetElemIndex(node.Value.GetType()))),
		self.buildStructIndex(ptr, 1, true),
	)
	if load {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenUnionTypeJudgment(node *hir.UnionTypeJudgment) mir.Value {
	utMean := node.Value.GetType().(*hir.UnionType)
	ut := self.codegenUnionType(utMean)
	typeIndex := self.buildStructIndex(self.codegenExpr(node.Value, false), 1, false)
	return self.builder.BuildCmp(mir.CmpKindEQ, typeIndex, mir.NewInt(ut.Elems()[1].(mir.IntType), int64(utMean.GetElemIndex(node.Type))))
}

func (self *CodeGenerator) codegenUnUnion(node *hir.UnUnion) mir.Value {
	value := self.codegenExpr(node.Value, false)
	elemPtr := self.buildStructIndex(value, 0, true)
	elemPtr = self.builder.BuildPtrToPtr(elemPtr, self.ctx.NewPtrType(self.codegenType(node.GetType())))
	return self.builder.BuildLoad(elemPtr)
}

func (self *CodeGenerator) codegenWrapWithNull(node *hir.WrapWithNull, load bool) mir.Value {
	return self.codegenExpr(node.Value, load)
}

func (self *CodeGenerator) codegenCheckNull(node *hir.CheckNull) mir.Value {
	name := "sim_runtime_check_null"
	ptrType := self.ctx.NewPtrType(self.ctx.U8())
	ft := self.ctx.NewFuncType(ptrType, ptrType)
	f, ok := self.module.NamedFunction(name)
	if !ok {
		f = self.module.NewFunction(name, ft)
	}

	ptr := self.codegenExpr(node.Value, true)
	return self.builder.BuildCall(f, ptr)
}
