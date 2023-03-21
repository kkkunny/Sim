package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
)

var (
	t_bool llvm.Type
	t_size llvm.Type

	v_true  llvm.Value
	v_false llvm.Value
)

func (self CodeGenerator) init() {
	t_bool = self.ctx.Int8Type()
	t_size = self.targetData.IntPtrType()

	v_true = llvm.ConstInt(t_bool, 1, true)
	v_false = llvm.ConstInt(t_bool, 0, true)
}

func (self CodeGenerator) createAllocaHirType(t hir.Type) llvm.Value {
	alloca := self.builder.CreateAlloca(self.codegenType(t), "")
	alloca.SetAlignment(int(t.Align()))
	return alloca
}

func (self *CodeGenerator) createArrayIndex(v llvm.Value, i llvm.Value, getValue bool, align int) llvm.Value {
	if v.Type().TypeKind() == llvm.PointerTypeKind {
		value := self.builder.CreateInBoundsGEP(v, []llvm.Value{llvm.ConstInt(t_size, 0, false), i}, "")
		if getValue {
			value = self.builder.CreateLoad(value, "")
			value.SetAlignment(align)
		}
		return value
	} else {
		return self.builder.CreateExtractElement(v, i, "")
	}
}

func (self *CodeGenerator) createPointerIndex(v llvm.Value, i llvm.Value, getValue bool, align int) llvm.Value {
	value := self.builder.CreateInBoundsGEP(v, []llvm.Value{i}, "")
	if getValue {
		value = self.builder.CreateLoad(value, "")
		value.SetAlignment(align)
	}
	return value
}

func (self *CodeGenerator) createStructIndex(v llvm.Value, i uint, getValue bool, align int) llvm.Value {
	if v.Type().TypeKind() == llvm.PointerTypeKind {
		value := self.builder.CreateStructGEP(v, int(i), "")
		if getValue {
			value = self.builder.CreateLoad(value, "")
			value.SetAlignment(align)
		}
		return value
	} else {
		return self.builder.CreateExtractValue(v, int(i), "")
	}
}
