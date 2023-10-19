package codegen

import (
	"github.com/kkkunny/llvm"

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
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *mean.Integer) llvm.Value {
	t := self.codegenType(node.GetType())
	return llvm.ConstIntFromString(t, node.Value.String(), 10)
}

func (self *CodeGenerator) codegenFloat(node *mean.Float) llvm.Value {
	t := self.codegenType(node.GetType())
	return llvm.ConstFloatFromString(t, node.Value.String())
}

func (self *CodeGenerator) codegenBool(node *mean.Boolean) llvm.Value {
	t := self.codegenType(node.GetType())
	if node.Value {
		return llvm.ConstInt(t, 1, false)
	} else {
		return llvm.ConstInt(t, 0, false)
	}
}

func (self *CodeGenerator) codegenBinary(node mean.Binary) llvm.Value {
	left, right := self.codegenExpr(node.GetLeft()), self.codegenExpr(node.GetRight())

	switch node.(type) {
	case *mean.IntAndInt:
		return self.builder.CreateAnd(left, right, "")
	case *mean.IntOrInt:
		return self.builder.CreateOr(left, right, "")
	case *mean.IntXorInt:
		return self.builder.CreateXor(left, right, "")
	case *mean.NumAddNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWAdd(left, right, "")
		case *mean.UintType:
			return self.builder.CreateNUWAdd(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFAdd(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumSubNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWSub(left, right, "")
		case *mean.UintType:
			return self.builder.CreateNUWSub(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFSub(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumMulNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWMul(left, right, "")
		case *mean.UintType:
			return self.builder.CreateNUWMul(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFMul(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumDivNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSDiv(left, right, "")
		case *mean.UintType:
			return self.builder.CreateUDiv(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFDiv(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumRemNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSRem(left, right, "")
		case *mean.UintType:
			return self.builder.CreateURem(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFRem(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumLtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSLT, left, right, "")
		case *mean.UintType:
			return self.builder.CreateICmp(llvm.IntULT, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOLT, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumGtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSGT, left, right, "")
		case *mean.UintType:
			return self.builder.CreateICmp(llvm.IntUGT, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOGT, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumLeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSLE, left, right, "")
		case *mean.UintType:
			return self.builder.CreateICmp(llvm.IntULE, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOLE, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumGeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSGE, left, right, "")
		case *mean.UintType:
			return self.builder.CreateICmp(llvm.IntUGE, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOGE, left, right, "")
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
			return self.builder.CreateNSWNeg(value, "")
		case *mean.FloatType:
			return self.builder.CreateFNeg(value, "")
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
		f := self.module.NamedFunction(identNode.Name)
		return f
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(node *mean.Call) llvm.Value {
	t := self.codegenType(node.Func.GetType())
	f := self.codegenExpr(node.Func)
	return self.builder.CreateCall(t, f, nil, "")
}

func (self *CodeGenerator) codegenCovert(node mean.Covert) llvm.Value {
	ft := node.GetFrom().GetType()
	from := self.codegenExpr(node.GetFrom())
	to := self.codegenType(node.GetType())

	switch covertNode := node.(type) {
	case *mean.Num2Num:
		switch {
		case mean.TypeIs[mean.IntType](ft) && mean.TypeIs[mean.IntType](covertNode.To):
			ift, itt := ft.(mean.IntType), covertNode.To.(mean.IntType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateSExt(from, to, "")
			} else if ifb > itb {
				return self.builder.CreateTrunc(from, to, "")
			} else {
				return from
			}
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			ift, itt := ft.(*mean.FloatType), covertNode.To.(*mean.FloatType)
			if ifb, itb := ift.GetBits(), itt.GetBits(); ifb < itb {
				return self.builder.CreateFPExt(from, to, "")
			} else if ifb > itb {
				return self.builder.CreateFPTrunc(from, to, "")
			} else {
				return from
			}
		case mean.TypeIs[*mean.SintType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			return self.builder.CreateSIToFP(from, to, "")
		case mean.TypeIs[*mean.UintType](ft) && mean.TypeIs[*mean.FloatType](covertNode.To):
			return self.builder.CreateUIToFP(from, to, "")
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.SintType](covertNode.To):
			return self.builder.CreateFPToSI(from, to, "")
		case mean.TypeIs[*mean.FloatType](ft) && mean.TypeIs[*mean.UintType](covertNode.To):
			return self.builder.CreateFPToUI(from, to, "")
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}
