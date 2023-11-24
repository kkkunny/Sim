package codegen

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
		return self.codegenZero(exprNode)
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
	case *mean.FuncDef:
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
	f := self.codegenExpr(node.Func, true)
	args := lo.Map(node.Args, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item, true)
	})
	return self.buildCall(load, self.codegenFuncType(node.Func.GetType().(*mean.FuncType)), f, args...)
}

func (self *CodeGenerator) codegenCovert(node mean.Covert) llvm.Value {
	ft := node.GetFrom().GetType()
	from := self.codegenExpr(node.GetFrom(), true)
	to := self.codegenType(node.GetType())

	switch covertNode := node.(type) {
	case *mean.Num2Num:
		switch {
		case mean.TypeIs[*mean.SintType](ft) && mean.TypeIs[mean.IntType](covertNode.To):
			ift, itt := ft.(*mean.SintType), covertNode.To.(mean.IntType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateSExt("", from, to.(llvm.IntegerType))
			} else if ifb > itb {
				return self.builder.CreateTrunc("", from, to.(llvm.IntegerType))
			} else {
				return from
			}
		case mean.TypeIs[*mean.UintType](ft) && mean.TypeIs[mean.IntType](covertNode.To):
			ift, itt := ft.(*mean.UintType), covertNode.To.(mean.IntType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateZExt("", from, to.(llvm.IntegerType))
			} else if ifb > itb {
				return self.builder.CreateTrunc("", from, to.(llvm.IntegerType))
			} else {
				return from
			}
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			ift, itt := ft.(*mean.FloatType), covertNode.To.(*mean.FloatType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateFPExt("", from, to.(llvm.FloatType))
			} else if ifb > itb {
				return self.builder.CreateFPTrunc("", from, to.(llvm.FloatType))
			} else {
				return from
			}
		case mean.TypeIs[*mean.SintType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			return self.builder.CreateSIToFP("", from, to.(llvm.FloatType))
		case mean.TypeIs[*mean.UintType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			return self.builder.CreateUIToFP("", from, to.(llvm.FloatType))
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.SintType](covertNode.To):
			return self.builder.CreateFPToSI("", from, to.(llvm.IntegerType))
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.UintType](covertNode.To):
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

func (self *CodeGenerator) codegenZero(node *mean.Zero) llvm.Constant {
	t := self.codegenType(node.GetType())
	switch node.GetType().(type) {
	case mean.IntType:
		return self.ctx.ConstInteger(t.(llvm.IntegerType), 0)
	case *mean.FloatType:
		return self.ctx.ConstFloat(t.(llvm.FloatType), 0)
	case *mean.BoolType:
		return self.ctx.ConstInteger(t.(llvm.IntegerType), 0)
	case *mean.ArrayType:
		return self.ctx.ConstAggregateZero(t.(llvm.ArrayType))
	case *mean.StructType:
		return self.ctx.ConstAggregateZero(t.(llvm.StructType))
	case *mean.StringType:
		return self.ctx.ConstAggregateZero(t.(llvm.StructType))
	case *mean.UnionType:
		return self.ctx.ConstAggregateZero(t.(llvm.StructType))
	case *mean.PtrType:
		return self.ctx.ConstNull(t)
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
	st := self.codegenStringType(node.GetType().(*mean.StringType))
	ptr := self.strings.Get(node.Value)
	if ptr == nil {
		data := self.ctx.ConstString(node.Value)
		dataPtr := self.module.NewGlobal("", data.Type())
		dataPtr.SetGlobalConstant(true)
		dataPtr.SetLinkage(llvm.PrivateLinkage)
		dataPtr.SetUnnamedAddress(llvm.GlobalUnnamedAddr)
		dataPtr.SetInitializer(data)
		global := self.module.NewGlobal("", st)
		ptr = &global
		ptr.SetGlobalConstant(true)
		ptr.SetLinkage(llvm.PrivateLinkage)
		ptr.SetUnnamedAddress(llvm.GlobalUnnamedAddr)
		ptr.SetInitializer(self.ctx.ConstNamedStruct(st, dataPtr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(len(node.Value)))))
		self.strings.Set(node.Value, ptr)
	}
	return self.builder.CreateLoad("", st, ptr)
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
