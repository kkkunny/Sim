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
		case *mean.FloatType:
			return self.builder.CreateFAdd(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumSubNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWSub(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFSub(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumMulNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWMul(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFMul(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumDivNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSDiv(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFDiv(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumRemNum:
		switch node.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSRem(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFRem(left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumLtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSLT, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOLT, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumGtNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSGT, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOGT, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumLeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSLE, left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFCmp(llvm.FloatOLE, left, right, "")
		default:
			panic("unreachable")
		}
	case *mean.NumGeNum:
		switch node.GetLeft().GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateICmp(llvm.IntSGE, left, right, "")
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
