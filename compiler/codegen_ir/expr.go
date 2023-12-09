package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenExpr(node mean.Expr, load bool) llvm.Value {
	switch exprNode := node.(type) {
	case *mean.Integer:
		return self.codegenInteger(exprNode)
	case *mean.Float:
		return self.codegenFloat(exprNode)
	case *mean.Boolean:
		return self.codegenBool(exprNode)
	case *mean.Assign:
		self.codegenAssign(exprNode)
		return self.ctx.ConstNull(self.ctx.IntPtrType(self.target))
	case mean.Binary:
		return self.codegenBinary(exprNode)
	case mean.Unary:
		return self.codegenUnary(exprNode, load)
	case mean.Ident:
		return self.codegenIdent(exprNode, load)
	case *mean.Call:
		return self.codegenCall(exprNode, load)
	case mean.Covert:
		return self.codegenCovert(exprNode)
	case *mean.Array:
		return self.codegenArray(exprNode, load)
	case *mean.Index:
		return self.codegenIndex(exprNode, load)
	case *mean.Tuple:
		return self.codegenTuple(exprNode, load)
	case *mean.Extract:
		return self.codegenExtract(exprNode, load)
	case *mean.Zero:
		return self.codegenZero(exprNode.GetType())
	case *mean.Struct:
		return self.codegenStruct(exprNode, load)
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
		return self.codegenCheckNull(exprNode, load)
	case *mean.MethodDef:
		// TODO: 闭包
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *mean.Integer) llvm.Value {
	t := self.codegenIntType(node.Type)
	return self.ctx.ConstIntegerFromString(t, node.Value.String(), 10)
}

func (self *CodeGenerator) codegenFloat(node *mean.Float) llvm.Value {
	t := self.codegenFloatType(node.Type)
	return self.ctx.ConstFloatFromString(t, node.Value.String())
}

func (self *CodeGenerator) codegenBool(node *mean.Boolean) llvm.Value {
	t := self.codegenType(node.GetType())
	if node.Value {
		return self.ctx.ConstInteger(t.(llvm.IntegerType), 1)
	} else {
		return self.ctx.ConstInteger(t.(llvm.IntegerType), 0)
	}
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
		self.builder.CreateStore(right, left)
	}
}

func (self *CodeGenerator) codegenBinary(node mean.Binary) llvm.Value {
	left, right := self.codegenExpr(node.GetLeft(), true), self.codegenExpr(node.GetRight(), true)
	switch node.(type) {
	case *mean.IntAndInt, *mean.BoolAndBool:
		return self.builder.CreateAnd("", left, right)
	case *mean.IntOrInt, *mean.BoolOrBool:
		return self.builder.CreateOr("", left, right)
	case *mean.IntXorInt:
		return self.builder.CreateXor("", left, right)
	case *mean.IntShlInt:
		return self.builder.CreateShl("", left, right)
	case *mean.IntShrInt:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateAShr("", left, right)
		case *mean.UintType:
			return self.builder.CreateLShr("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumAddNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSAdd("", left, right)
		case *mean.UintType:
			return self.builder.CreateUAdd("", left, right)
		case *mean.FloatType:
			return self.builder.CreateFAdd("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumSubNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSSub("", left, right)
		case *mean.UintType:
			return self.builder.CreateUSub("", left, right)
		case *mean.FloatType:
			return self.builder.CreateFSub("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumMulNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSMul("", left, right)
		case *mean.UintType:
			return self.builder.CreateUMul("", left, right)
		case *mean.FloatType:
			return self.builder.CreateFMul("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumDivNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSDiv("", left, right)
		case *mean.UintType:
			return self.builder.CreateUDiv("", left, right)
		case *mean.FloatType:
			return self.builder.CreateFDiv("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumRemNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSRem("", left, right)
		case *mean.UintType:
			return self.builder.CreateURem("", left, right)
		case *mean.FloatType:
			return self.builder.CreateFRem("", left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumLtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateIntCmp("", llvm.IntSLT, left, right)
		case *mean.UintType:
			return self.builder.CreateIntCmp("", llvm.IntULT, left, right)
		case *mean.FloatType:
			return self.builder.CreateFloatCmp("", llvm.FloatOLT, left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumGtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateIntCmp("", llvm.IntSGT, left, right)
		case *mean.UintType:
			return self.builder.CreateIntCmp("", llvm.IntUGT, left, right)
		case *mean.FloatType:
			return self.builder.CreateFloatCmp("", llvm.FloatOGT, left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumLeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateIntCmp("", llvm.IntSLE, left, right)
		case *mean.UintType:
			return self.builder.CreateIntCmp("", llvm.IntULE, left, right)
		case *mean.FloatType:
			return self.builder.CreateFloatCmp("", llvm.FloatOLE, left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumGeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateIntCmp("", llvm.IntSGE, left, right)
		case *mean.UintType:
			return self.builder.CreateIntCmp("", llvm.IntUGE, left, right)
		case *mean.FloatType:
			return self.builder.CreateFloatCmp("", llvm.FloatOGE, left, right)
		default:
			panic("unreachable")
		}
	case *mean.NumEqNum, *mean.BoolEqBool, *mean.FuncEqFunc, *mean.ArrayEqArray, *mean.StructEqStruct, *mean.TupleEqTuple, *mean.StringEqString, *mean.UnionEqUnion:
		return self.buildEqual(node.GetLeft().GetType(), left, right, false)
	case *mean.NumNeNum, *mean.BoolNeBool, *mean.FuncNeFunc, *mean.ArrayNeArray, *mean.StructNeStruct, *mean.TupleNeTuple, *mean.StringNeString, *mean.UnionNeUnion:
		return self.buildEqual(node.GetLeft().GetType(), left, right, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(node mean.Unary, load bool) llvm.Value {
	switch node.(type) {
	case *mean.NumNegate:
		value := self.codegenExpr(node.GetValue(), true)
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSNeg("", value)
		case *mean.FloatType:
			return self.builder.CreateFNeg("", value)
		default:
			panic("unreachable")
		}
	case *mean.IntBitNegate, *mean.BoolNegate:
		value := self.codegenExpr(node.GetValue(), true)
		return self.builder.CreateNot("", value)
	case *mean.GetPtr:
		return self.codegenExpr(node.GetValue(), false)
	case *mean.GetValue:
		ptr := self.codegenExpr(node.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.CreateLoad("", self.codegenType(node.GetType()), ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(node mean.Ident, load bool) llvm.Value {
	switch identNode := node.(type) {
	case *mean.FuncDef,*mean.GenericFuncInstance:
		return self.values[identNode].(llvm.Function)
	case *mean.Param, *mean.Variable:
		p := self.values[identNode]
		t := self.codegenType(node.GetType())
		if !load {
			return p
		}
		return self.builder.CreateLoad("", t, p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(node *mean.Call, load bool) llvm.Value {
	if method, ok := node.Func.(*mean.Method); ok{
		ft := self.codegenFuncType(method.Define.GetFuncType())
		f := self.values[method.Define]
		selfParam := self.codegenExpr(method.Self, true)
		args := lo.Map(node.Args, func(item mean.Expr, index int) llvm.Value {
			return self.codegenExpr(item, true)
		})
		return self.buildCall(load, ft, f, append([]llvm.Value{selfParam}, args...)...)
	}else if method, ok := node.Func.(*mean.TraitMethod); ok{
		return self.codegenTraitMethodCall(method, node.Args)
	} else{
		ft := self.codegenFuncType(node.Func.GetType().(*mean.FuncType))
		f := self.codegenExpr(node.Func, true)
		args := lo.Map(node.Args, func(item mean.Expr, index int) llvm.Value {
			return self.codegenExpr(item, true)
		})
		return self.buildCall(load, ft, f, args...)
	}
}

func (self *CodeGenerator) codegenCovert(node mean.Covert) llvm.Value {
	ft := node.GetFrom().GetType()
	from := self.codegenExpr(node.GetFrom(), true)
	to := self.codegenType(node.GetType())

	switch covertNode := node.(type) {
	case *mean.Num2Num:
		switch {
		case stlbasic.Is[*mean.SintType](ft) && stlbasic.Is[mean.IntType](covertNode.To):
			fs, ts := from.Type().(llvm.IntegerType).Bits(), to.(llvm.IntegerType).Bits()
			if fs < ts {
				return self.builder.CreateSExt("", from, to.(llvm.IntegerType))
			} else if fs > ts {
				return self.builder.CreateTrunc("", from, to.(llvm.IntegerType))
			} else {
				return from
			}
		case stlbasic.Is[*mean.UintType](ft) && stlbasic.Is[mean.IntType](covertNode.To):
			fs, ts := from.Type().(llvm.IntegerType).Bits(), to.(llvm.IntegerType).Bits()
			if fs < ts {
				return self.builder.CreateZExt("", from, to.(llvm.IntegerType))
			} else if fs > ts {
				return self.builder.CreateTrunc("", from, to.(llvm.IntegerType))
			} else {
				return from
			}
		case stlbasic.Is[*mean.FloatType](ft) && stlbasic.Is[*mean.FloatType](covertNode.To):
			ift, itt := ft.(*mean.FloatType), covertNode.To.(*mean.FloatType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateFPExt("", from, to.(llvm.FloatType))
			} else if ifb > itb {
				return self.builder.CreateFPTrunc("", from, to.(llvm.FloatType))
			} else {
				return from
			}
		case stlbasic.Is[*mean.SintType](ft) && stlbasic.Is[*mean.FloatType](covertNode.To):
			return self.builder.CreateSIToFP("", from, to.(llvm.FloatType))
		case stlbasic.Is[*mean.UintType](ft) && stlbasic.Is[*mean.FloatType](covertNode.To):
			return self.builder.CreateUIToFP("", from, to.(llvm.FloatType))
		case stlbasic.Is[*mean.FloatType](ft) && stlbasic.Is[*mean.SintType](covertNode.To):
			return self.builder.CreateFPToSI("", from, to.(llvm.IntegerType))
		case stlbasic.Is[*mean.FloatType](ft) && stlbasic.Is[*mean.UintType](covertNode.To):
			return self.builder.CreateFPToUI("", from, to.(llvm.IntegerType))
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(node *mean.Array, load bool) llvm.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item, true)
	})
	t := self.codegenType(node.Type).(llvm.ArrayType)
	ptr := self.builder.CreateAlloca("", t)
	for i, elem := range elems {
		self.builder.CreateStore(elem, self.buildArrayIndexWith(t, ptr, uint(i), false))
	}
	if !load {
		return ptr
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenIndex(node *mean.Index, load bool) llvm.Value {
	// TODO: 运行时异常：超出索引下标
	t := self.codegenType(node.From.GetType()).(llvm.ArrayType)
	from := self.codegenExpr(node.From, false)
	if _, ok := from.Type().(llvm.PointerType); !ok {
		ptr := self.builder.CreateAlloca("", t)
		self.builder.CreateStore(from, ptr)
		from = ptr
	}
	return self.buildArrayIndex(t, from, self.codegenExpr(node.Index, true), load)
}

func (self *CodeGenerator) codegenTuple(node *mean.Tuple, load bool) llvm.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item, true)
	})
	t := self.codegenType(node.GetType()).(llvm.StructType)
	ptr := self.builder.CreateAlloca("", t)
	for i, elem := range elems {
		self.builder.CreateStore(elem, self.buildStructIndex(t, ptr, uint(i), false))
	}
	if !load {
		return ptr
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenExtract(node *mean.Extract, load bool) llvm.Value {
	t := self.codegenType(node.From.GetType()).(llvm.StructType)
	from := self.codegenExpr(node.From, false)
	if _, ok := from.Type().(llvm.PointerType); !ok {
		ptr := self.builder.CreateAlloca("", t)
		self.builder.CreateStore(from, ptr)
		from = ptr
	}
	return self.buildStructIndex(t, from, node.Index, load)
}

// TODO: 复杂类型default值
func (self *CodeGenerator) codegenZero(tNode mean.Type) llvm.Value {
	switch ttNode := tNode.(type) {
	case mean.IntType:
		return self.ctx.ConstInteger(self.codegenIntType(ttNode), 0)
	case *mean.FloatType:
		return self.ctx.ConstFloat(self.codegenFloatType(ttNode), 0)
	case *mean.BoolType:
		return self.ctx.ConstInteger(self.codegenBoolType(), 0)
	case *mean.ArrayType:
		return self.ctx.ConstAggregateZero(self.codegenArrayType(ttNode))
	case *mean.StructType:
		return self.ctx.ConstAggregateZero(self.codegenStructType(ttNode))
	case *mean.StringType:
		return self.ctx.ConstAggregateZero(self.codegenStringType())
	case *mean.UnionType:
		return self.ctx.ConstAggregateZero(self.codegenUnionType(ttNode))
	case *mean.PtrType:
		return self.ctx.ConstNull(self.codegenPtrType(ttNode))
	case *mean.RefType:
		return self.ctx.ConstNull(self.codegenRefType(ttNode))
	case *mean.GenericParam:
		return self.codegenCall(&mean.Call{Func: &mean.TraitMethod{
			Type: ttNode,
			Name: "default",
		}}, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(node *mean.Struct, load bool) llvm.Value {
	fields := lo.Map(node.Fields, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item, true)
	})
	t := self.codegenType(node.GetType()).(llvm.StructType)
	ptr := self.builder.CreateAlloca("", t)
	for i, field := range fields {
		self.builder.CreateStore(field, self.buildStructIndex(t, ptr, uint(i), false))
	}
	if !load {
		return ptr
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenField(node *mean.Field, load bool) llvm.Value {
	t := self.codegenType(node.From.GetType()).(llvm.StructType)
	from := self.codegenExpr(node.From, false)
	if _, ok := from.Type().(llvm.PointerType); !ok {
		ptr := self.builder.CreateAlloca("", t)
		self.builder.CreateStore(from, ptr)
		from = ptr
	}
	return self.buildStructIndex(t, from, node.Index, load)
}

func (self *CodeGenerator) codegenString(node *mean.String) llvm.Value {
	st := self.codegenStringType()
	ptr := self.strings.Get(node.Value)
	if ptr == nil {
		data := self.ctx.ConstString(node.Value)
		dataPtr := self.module.NewGlobal("", data.Type())
		dataPtr.SetGlobalConstant(true)
		dataPtr.SetLinkage(llvm.PrivateLinkage)
		dataPtr.SetUnnamedAddress(llvm.GlobalUnnamedAddr)
		dataPtr.SetInitializer(data)
		self.strings.Set(node.Value, &dataPtr)
		ptr = &dataPtr
	}
	return self.ctx.ConstNamedStruct(st, *ptr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(len(node.Value))))
}

func (self *CodeGenerator) codegenUnion(node *mean.Union, load bool) llvm.Value {
	ut := self.codegenUnionType(node.Type)
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.CreateAlloca("", ut)
	self.builder.CreateStore(value, self.buildStructIndex(ut, ptr, 0, false))
	self.builder.CreateStore(self.ctx.ConstInteger(ut.GetElem(1).(llvm.IntegerType), int64(node.Type.GetElemIndex(node.Value.GetType()))), self.buildStructIndex(ut, ptr, 1, false))
	if load {
		return self.builder.CreateLoad("", ut, ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenUnionTypeJudgment(node *mean.UnionTypeJudgment) llvm.Value {
	utMean := node.Value.GetType().(*mean.UnionType)
	ut := self.codegenUnionType(utMean)
	typeIndex := self.buildStructIndex(ut, self.codegenExpr(node.Value, false), 1, true)
	return self.builder.CreateIntCmp("", llvm.IntEQ, typeIndex, self.ctx.ConstInteger(ut.GetElem(1).(llvm.IntegerType), int64(utMean.GetElemIndex(node.Type))))
}

func (self *CodeGenerator) codegenUnUnion(node *mean.UnUnion) llvm.Value {
	t := self.codegenType(node.GetType())
	value := self.codegenExpr(node.Value, false)
	if stlbasic.Is[llvm.PointerType](value.Type()) {
		return self.builder.CreateLoad("", t, value)
	}
	elem := self.buildStructIndex(value.Type().(llvm.StructType), value, 0, true)
	ptr := self.builder.CreateAlloca("", elem.Type())
	self.builder.CreateStore(elem, ptr)
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenWrapWithNull(node *mean.WrapWithNull, load bool) llvm.Value {
	return self.codegenExpr(node.Value, load)
}

func (self *CodeGenerator) codegenCheckNull(node *mean.CheckNull, load bool) llvm.Value {
	name := "sim_runtime_check_null"
	ptrType := self.ctx.PointerType(self.ctx.IntegerType(8))
	ft := self.ctx.FunctionType(false, ptrType, ptrType)
	f, ok := self.module.GetFunction(name)
	if !ok {
		f = self.newFunction(name, ft)
	}

	ptr := self.codegenExpr(node.Value, true)
	return self.buildCall(load, ft, f, ptr)
}

func (self *CodeGenerator) codegenTraitMethodCall(node *mean.TraitMethod, args []mean.Expr)llvm.Value{
	autTypeNodeObj := self.genericParams.Get(node.Type)
	switch autTypeNode:=autTypeNodeObj.(type) {
	case *mean.StructType:
		method := autTypeNode.GetImplMethod(node.Name, node.GetType().(*mean.FuncType))
		ft := self.codegenFuncType(method.GetFuncType())
		f := self.values[method]
		var selfParam llvm.Value
		if selfNode, ok := node.Value.Value(); ok{
			selfParam = self.codegenExpr(selfNode, true)
		}else{
			selfParam = self.ctx.ConstAggregateZero(self.codegenStructType(method.Scope))
		}
		args := lo.Map(args, func(item mean.Expr, index int) llvm.Value {
			return self.codegenExpr(item, true)
		})
		return self.buildCall(true, ft, f, append([]llvm.Value{selfParam}, args...)...)
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