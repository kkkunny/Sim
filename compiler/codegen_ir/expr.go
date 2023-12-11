package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenExpr(node mean.Expr, load bool) mir.Value {
	switch exprNode := node.(type) {
	case *mean.Integer:
		return self.codegenInteger(exprNode)
	case *mean.Float:
		return self.codegenFloat(exprNode)
	case *mean.Boolean:
		return self.codegenBool(exprNode)
	case *mean.Assign:
		self.codegenAssign(exprNode)
		return nil
	case mean.Binary:
		return self.codegenBinary(exprNode)
	case mean.Unary:
		return self.codegenUnary(exprNode, load)
	case mean.Ident:
		return self.codegenIdent(exprNode, load)
	case *mean.Call:
		return self.codegenCall(exprNode)
	case mean.Covert:
		return self.codegenCovert(exprNode)
	case *mean.Array:
		return self.codegenArray(exprNode)
	case *mean.Index:
		return self.codegenIndex(exprNode, load)
	case *mean.Tuple:
		return self.codegenTuple(exprNode)
	case *mean.Extract:
		return self.codegenExtract(exprNode, load)
	case *mean.Zero:
		return self.codegenZero(exprNode.GetType())
	case *mean.Struct:
		return self.codegenStruct(exprNode)
	case *mean.Field:
		return self.codegenField(exprNode, load)
	case *mean.String:
		return self.codegenString(exprNode)
	case *mean.Union:
		return self.codegenUnion(exprNode, load)
	case *mean.UnionTypeJudgment:
		return self.codegenUnionTypeJudgment(exprNode)
	case *mean.UnUnion:
		return self.codegenUnUnion(exprNode)
	case *mean.WrapWithNull:
		return self.codegenWrapWithNull(exprNode, load)
	case *mean.CheckNull:
		return self.codegenCheckNull(exprNode)
	case *mean.MethodDef:
		// TODO: 闭包
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *mean.Integer) mir.Int {
	return mir.NewInt(self.codegenIntType(node.Type), node.Value.Int64())
}

func (self *CodeGenerator) codegenFloat(node *mean.Float) *mir.Float {
	v, _ := node.Value.Float64()
	return mir.NewFloat(self.codegenFloatType(node.Type), v)
}

func (self *CodeGenerator) codegenBool(node *mean.Boolean) *mir.Uint {
	return mir.Bool(self.ctx, node.Value)
}

func (self *CodeGenerator) codegenAssign(node *mean.Assign) {
	if lpack, ok := node.Left.(*mean.Tuple); ok {
		// 解包
		if rpack, ok := node.Right.(*mean.Tuple); ok {
			for i, le := range lpack.Elems {
				re := rpack.Elems[i]
				self.codegenAssign(&mean.Assign{
					Left:  le,
					Right: re,
				})
			}
		} else {
			for i, le := range lpack.Elems {
				self.codegenAssign(&mean.Assign{
					Left: le,
					Right: &mean.Extract{
						From:  node.Right,
						Index: uint(i),
					},
				})
			}
		}
	} else {
		left, right := self.codegenExpr(node.GetLeft(), false), self.codegenExpr(node.GetRight(), true)
		self.builder.BuildStore(right, left)
	}
}

func (self *CodeGenerator) codegenBinary(node mean.Binary) mir.Value {
	left, right := self.codegenExpr(node.GetLeft(), true), self.codegenExpr(node.GetRight(), true)
	switch node.(type) {
	case *mean.IntAndInt, *mean.BoolAndBool:
		return self.builder.BuildAnd(left, right)
	case *mean.IntOrInt, *mean.BoolOrBool:
		return self.builder.BuildOr(left, right)
	case *mean.IntXorInt:
		return self.builder.BuildXor(left, right)
	case *mean.IntShlInt:
		return self.builder.BuildShl(left, right)
	case *mean.IntShrInt:
		return self.builder.BuildShr(left, right)
	case *mean.NumAddNum:
		return self.builder.BuildAdd(left, right)
	case *mean.NumSubNum:
		return self.builder.BuildSub(left, right)
	case *mean.NumMulNum:
		return self.builder.BuildMul(left, right)
	case *mean.NumDivNum:
		return self.builder.BuildDiv(left, right)
	case *mean.NumRemNum:
		return self.builder.BuildRem(left, right)
	case *mean.NumLtNum:
		return self.builder.BuildCmp(mir.CmpKindLT, left, right)
	case *mean.NumGtNum:
		return self.builder.BuildCmp(mir.CmpKindGT, left, right)
	case *mean.NumLeNum:
		return self.builder.BuildCmp(mir.CmpKindLE, left, right)
	case *mean.NumGeNum:
		return self.builder.BuildCmp(mir.CmpKindGE, left, right)
	case *mean.NumEqNum, *mean.BoolEqBool, *mean.FuncEqFunc, *mean.ArrayEqArray, *mean.StructEqStruct, *mean.TupleEqTuple, *mean.StringEqString, *mean.UnionEqUnion:
		return self.buildEqual(node.GetLeft().GetType(), left, right, false)
	case *mean.NumNeNum, *mean.BoolNeBool, *mean.FuncNeFunc, *mean.ArrayNeArray, *mean.StructNeStruct, *mean.TupleNeTuple, *mean.StringNeString, *mean.UnionNeUnion:
		return self.buildEqual(node.GetLeft().GetType(), left, right, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(node mean.Unary, load bool) mir.Value {
	switch node.(type) {
	case *mean.NumNegate:
		return self.builder.BuildNeg(self.codegenExpr(node.GetValue(), true))
	case *mean.IntBitNegate, *mean.BoolNegate:
		return self.builder.BuildNot(self.codegenExpr(node.GetValue(), true))
	case *mean.GetPtr:
		return self.codegenExpr(node.GetValue(), false)
	case *mean.GetValue:
		ptr := self.codegenExpr(node.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(node mean.Ident, load bool) mir.Value {
	switch identNode := node.(type) {
	case *mean.FuncDef,*mean.GenericFuncInstance:
		return self.values.Get(identNode).(*mir.Function)
	case *mean.Param, *mean.Variable:
		p := self.values.Get(identNode)
		if !load {
			return p
		}
		return self.builder.BuildLoad(p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(node *mean.Call) mir.Value {
	if method, ok := node.Func.(*mean.Method); ok{
		f := self.values.Get(method.Define)
		selfParam := self.codegenExpr(method.Self, true)
		args := lo.Map(node.Args, func(item mean.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, append([]mir.Value{selfParam}, args...)...)
	}else if method, ok := node.Func.(*mean.TraitMethod); ok{
		return self.codegenTraitMethodCall(method, node.Args)
	} else{
		f := self.codegenExpr(node.Func, true)
		args := lo.Map(node.Args, func(item mean.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(node mean.Covert) mir.Value {
	from := self.codegenExpr(node.GetFrom(), true)
	to := self.codegenType(node.GetType())

	switch node.(type) {
	case *mean.Num2Num:
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(node *mean.Array) mir.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackArray(self.codegenArrayType(node.Type), elems...)
}

func (self *CodeGenerator) codegenIndex(node *mean.Index, load bool) mir.Value {
	// TODO: 运行时异常：超出索引下标
	from := self.codegenExpr(node.From, false)
	ptr := self.buildArrayIndex(from, self.codegenExpr(node.Index, true))
	if !load || (stlbasic.Is[*mir.ArrayIndex](ptr) && !ptr.(*mir.ArrayIndex).IsPtr()){
		return ptr
	}
	return self.builder.BuildLoad(ptr)
}

func (self *CodeGenerator) codegenTuple(node *mean.Tuple) mir.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenTupleType(node.GetType().(*mean.TupleType)), elems...)
}

func (self *CodeGenerator) codegenExtract(node *mean.Extract, load bool) mir.Value {
	from := self.codegenExpr(node.From, false)
	ptr := self.buildStructIndex(from, uint64(node.Index))
	if !load || (stlbasic.Is[*mir.StructIndex](ptr) && !ptr.(*mir.StructIndex).IsPtr()){
		return ptr
	}
	return self.builder.BuildLoad(ptr)
}

func (self *CodeGenerator) codegenZero(tNode mean.Type) mir.Value {
	switch ttNode := tNode.(type) {
	case mean.NumberType, *mean.BoolType, *mean.StringType, *mean.PtrType, *mean.RefType:
		return mir.NewZero(self.codegenType(ttNode))
	case *mean.ArrayType, *mean.StructType, *mean.UnionType:
		// TODO: 复杂类型default值
		return mir.NewZero(self.codegenType(ttNode))
	case *mean.GenericParam:
		return self.codegenCall(&mean.Call{Func: &mean.TraitMethod{
			Type: ttNode,
			Name: "default",
		}})
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(node *mean.Struct) mir.Value {
	fields := lo.Map(node.Fields, func(item mean.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenStructType(node.Type), fields...)
}

func (self *CodeGenerator) codegenField(node *mean.Field, load bool) mir.Value {
	from := self.codegenExpr(node.From, false)
	ptr := self.buildStructIndex(from, uint64(node.Index))
	if !load || (stlbasic.Is[*mir.StructIndex](ptr) && !ptr.(*mir.StructIndex).IsPtr()){
		return ptr
	}
	return self.builder.BuildLoad(ptr)
}

func (self *CodeGenerator) codegenString(node *mean.String) mir.Value {
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

func (self *CodeGenerator) codegenUnion(node *mean.Union, load bool) mir.Value {
	ut := self.codegenUnionType(node.Type)
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.BuildAllocFromStack(ut)
	self.builder.BuildStore(value, self.buildStructIndex(ptr, 0, true))
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(node.Type.GetElemIndex(node.Value.GetType()))),
		self.buildStructIndex(ptr, 1, true),
	)
	if load {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenUnionTypeJudgment(node *mean.UnionTypeJudgment) mir.Value {
	utMean := node.Value.GetType().(*mean.UnionType)
	ut := self.codegenUnionType(utMean)
	typeIndex := self.buildStructIndex(self.codegenExpr(node.Value, false), 1, false)
	return self.builder.BuildCmp(mir.CmpKindEQ, typeIndex, mir.NewInt(ut.Elems()[1].(mir.IntType), int64(utMean.GetElemIndex(node.Type))))
}

func (self *CodeGenerator) codegenUnUnion(node *mean.UnUnion) mir.Value {
	value := self.codegenExpr(node.Value, false)
	if stlbasic.Is[llvm.PointerType](value.Type()) {
		return self.builder.BuildLoad(value)
	}
	elem := self.buildStructIndex(value, 0, false)
	ptr := self.builder.BuildAllocFromStack(elem.Type())
	self.builder.BuildStore(elem, ptr)
	return self.builder.BuildLoad(ptr)
}

func (self *CodeGenerator) codegenWrapWithNull(node *mean.WrapWithNull, load bool) mir.Value {
	return self.codegenExpr(node.Value, load)
}

func (self *CodeGenerator) codegenCheckNull(node *mean.CheckNull) mir.Value {
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

func (self *CodeGenerator) codegenTraitMethodCall(node *mean.TraitMethod, args []mean.Expr)mir.Value{
	autTypeNodeObj := self.genericParams.Get(node.Type)
	switch autTypeNode:=autTypeNodeObj.(type) {
	case *mean.StructType:
		method := autTypeNode.GetImplMethod(node.Name, node.GetType().(*mean.FuncType))
		f := self.values.Get(method)
		var selfParam mir.Value
		if selfNode, ok := node.Value.Value(); ok{
			selfParam = self.codegenExpr(selfNode, true)
		}else{
			selfParam = mir.NewZero(self.codegenStructType(method.Scope))
		}
		args := lo.Map(args, func(item mean.Expr, index int) mir.Value {
			return self.codegenExpr(item, true)
		})
		return self.builder.BuildCall(f, append([]mir.Value{selfParam}, args...)...)
	case *mean.SintType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.UintType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.FloatType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.ArrayType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.TupleType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.PtrType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.RefType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.UnionType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.BoolType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	case *mean.StringType:
		switch node.Name {
		case "default":
			return self.codegenZero(autTypeNode)
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}
