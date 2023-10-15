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
	case *mean.Binary:
		return self.codegenBinary(exprNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *mean.Integer) llvm.Value {
	t := self.codegenType(node.GetType())
	return llvm.ConstInt(t, node.Value.Uint64(), node.Type.HasSign())
}

func (self *CodeGenerator) codegenFloat(node *mean.Float) llvm.Value {
	t := self.codegenType(node.GetType())
	v, _ := node.Value.Float64()
	return llvm.ConstFloat(t, v)
}

func (self *CodeGenerator) codegenBinary(node *mean.Binary) llvm.Value {
	left, right := self.codegenExpr(node.Left), self.codegenExpr(node.Right)

	switch node.Kind {
	case mean.BinaryAdd:
		switch node.Left.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWAdd(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFAdd(left, right, "")
		default:
			panic("unreachable")
		}
	case mean.BinarySub:
		switch node.Left.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWSub(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFSub(left, right, "")
		default:
			panic("unreachable")
		}
	case mean.BinaryMul:
		switch node.Left.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateNSWMul(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFMul(left, right, "")
		default:
			panic("unreachable")
		}
	case mean.BinaryDiv:
		switch node.Left.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSDiv(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFDiv(left, right, "")
		default:
			panic("unreachable")
		}
	case mean.BinaryRem:
		switch node.Left.GetType().(type) {
		case *mean.SintType:
			return self.builder.CreateSRem(left, right, "")
		case *mean.FloatType:
			return self.builder.CreateFRem(left, right, "")
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}
