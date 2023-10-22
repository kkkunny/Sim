package codegen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenExpr(node mean.Expr) llvm.Value {
	switch exprNode := node.(type) {
	case *mean.Integer:
		return self.codegenInteger(exprNode)
	case *mean.Float:
		return self.codegenFloat(exprNode)
	case *mean.Boolean:
		return self.codegenBool(exprNode)
	case mean.Binary:
		return self.codegenBinary(exprNode)
	case mean.Unary:
		return self.codegenUnary(exprNode)
	case mean.Ident:
		return self.codegenIdent(exprNode)
	case *mean.Call:
		return self.codegenCall(exprNode)
	case mean.Covert:
		return self.codegenCovert(exprNode)
	case *mean.Array:
		return self.codegenArray(exprNode)
	case *mean.Index:
		return self.codegenIndex(exprNode)
	case *mean.Tuple:
		return self.codegenTuple(exprNode)
	case *mean.Extract:
		return self.codegenExtract(exprNode)
	case *mean.Zero:
		return self.codegenZero(exprNode)
	case *mean.Struct:
		return self.codegenStruct(exprNode)
	case *mean.Field:
		return self.codegenField(exprNode)
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

func (self *CodeGenerator) codegenBinary(node mean.Binary) llvm.Value {
	left, right := self.codegenExpr(node.GetLeft()), self.codegenExpr(node.GetRight())

	switch node.(type) {
	case *mean.IntAndInt:
		return self.builder.CreateAnd("", left, right)
	case *mean.IntOrInt:
		return self.builder.CreateOr("", left, right)
	case *mean.IntXorInt:
		return self.builder.CreateXor("", left, right)
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
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(node mean.Unary) llvm.Value {
	value := self.codegenExpr(node.GetValue())

	switch node.(type) {
	case *mean.NumNegate:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSNeg("", value)
		case *mean.FloatType:
			return self.builder.CreateFNeg("", value)
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(node mean.Ident) llvm.Value {
	switch identNode := node.(type) {
	case *mean.FuncDef:
		f := self.module.GetFunction(identNode.Name)
		return f
	case *mean.Param:
		p := self.values[identNode]
		t := self.codegenType(node.GetType())
		return self.builder.CreateLoad("", t, p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(node *mean.Call) llvm.Value {
	ft := self.codegenFuncType(node.Func.GetType().(*mean.FuncType))
	f := self.codegenExpr(node.Func)
	args := lo.Map(node.Args, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item)
	})
	return self.builder.CreateCall("", ft, f, args...)
}

func (self *CodeGenerator) codegenCovert(node mean.Covert) llvm.Value {
	ft := node.GetFrom().GetType()
	from := self.codegenExpr(node.GetFrom())
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

func (self *CodeGenerator) codegenArray(node *mean.Array) llvm.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item)
	})
	t := self.codegenType(node.Type)
	ptr := self.builder.CreateAlloca("", t)
	for i, elem := range elems {
		ep := self.builder.CreateInBoundsGEP("", t, ptr, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(i)))
		self.builder.CreateStore(elem, ep)
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenIndex(node *mean.Index) llvm.Value {
	// TODO: 运行时异常：超出索引下标
	ft, et := self.codegenType(node.From.GetType()), self.codegenType(node.From.GetType().(*mean.ArrayType).Elem)
	ptr := self.builder.CreateAlloca("", ft)
	from := self.codegenExpr(node.From)
	self.builder.CreateStore(from, ptr)
	index := self.codegenExpr(node.Index)
	p := self.builder.CreateInBoundsGEP("", ft, ptr, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
	return self.builder.CreateLoad("", et, p)
}

func (self *CodeGenerator) codegenTuple(node *mean.Tuple) llvm.Value {
	elems := lo.Map(node.Elems, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item)
	})
	t := self.codegenTupleType(node.GetType().(*mean.TupleType))
	ptr := self.builder.CreateAlloca("", t)
	for i, elem := range elems {
		ep := self.builder.CreateStructGEP("", t, ptr, uint(i))
		self.builder.CreateStore(elem, ep)
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenExtract(node *mean.Extract) llvm.Value {
	ft, et := self.codegenType(node.From.GetType()), self.codegenType(node.GetType())
	ptr := self.builder.CreateAlloca("", ft)
	from := self.codegenExpr(node.From)
	self.builder.CreateStore(from, ptr)
	p := self.builder.CreateStructGEP("", ft, ptr, node.Index)
	return self.builder.CreateLoad("", et, p)
}

func (self *CodeGenerator) codegenZero(node *mean.Zero) llvm.Value {
	// TODO: 类型
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
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(node *mean.Struct) llvm.Value {
	fields := lo.Map(node.Fields, func(item mean.Expr, index int) llvm.Value {
		return self.codegenExpr(item)
	})
	t := self.codegenStructType(node.GetType().(*mean.StructType))
	ptr := self.builder.CreateAlloca("", t)
	for i, field := range fields {
		ep := self.builder.CreateStructGEP("", t, ptr, uint(i))
		self.builder.CreateStore(field, ep)
	}
	return self.builder.CreateLoad("", t, ptr)
}

func (self *CodeGenerator) codegenField(node *mean.Field) llvm.Value {
	ft, et := self.codegenType(node.From.GetType()), self.codegenType(node.GetType())
	ptr := self.builder.CreateAlloca("", ft)
	from := self.codegenExpr(node.From)
	self.builder.CreateStore(from, ptr)
	p := self.builder.CreateStructGEP("", ft, ptr, node.Index)
	return self.builder.CreateLoad("", et, p)
}
