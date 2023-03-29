package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/llvm"
)

// 值
func (self *CodeGenerator) codegenValue(ir mir.Value) llvm.Value {
	switch value := ir.(type) {
	case mir.Constant:
		return self.codegenConstant(value)
	case *mir.Function:
		return self.globals[value]
	default:
		return self.vars[value]
	}
}

// 常量
func (self *CodeGenerator) codegenConstant(ir mir.Constant) llvm.Value {
	switch value := ir.(type) {
	case *mir.Bool:
		if value.Value {
			return llvm.ConstInt(self.ctx.Int1Type(), 1, true)
		} else {
			return llvm.ConstInt(self.ctx.Int1Type(), 0, true)
		}
	case *mir.Sint:
		return llvm.ConstIntFromString(self.codegenType(value.Type), value.Value.String(), 10)
	case *mir.Uint:
		return llvm.ConstIntFromString(self.codegenType(value.Type), value.Value.String(), 10)
	case *mir.Float:
		return llvm.ConstFloatFromString(self.codegenType(value.Type), value.Value.String())
	case *mir.EmptyPtr, *mir.EmptyFunc:
		return llvm.ConstPointerNull(self.codegenType(value.GetType()))
	case *mir.EmptyArray, *mir.EmptyStruct, *mir.EmptyUnion:
		return llvm.ConstAggregateZero(self.codegenType(value.GetType()))
	case *mir.Array:
		elems := make([]llvm.Value, len(value.Elems))
		for i, e := range value.Elems {
			elems[i] = self.codegenConstant(e)
		}
		return llvm.ConstArray(self.codegenType(value.GetType()), elems)
	case *mir.Struct:
		elems := make([]llvm.Value, len(value.Elems))
		for i, e := range value.Elems {
			elems[i] = self.codegenConstant(e)
		}
		return self.ctx.ConstStruct(elems, false)
	case *mir.ArrayIndexConst:
		from := self.codegenConstant(value.From)
		index := llvm.ConstInt(self.targetData.IntPtrType(), uint64(value.Index), false)
		if value.IsPtr() {
			return llvm.ConstInBoundsGEP(from, []llvm.Value{index})
		}
		return llvm.ConstExtractElement(from, index)
	case *mir.Variable:
		return self.globals[value]
	default:
		panic("unreachable")
	}
}
