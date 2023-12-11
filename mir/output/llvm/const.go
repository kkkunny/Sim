package llvm

import (
	"github.com/kkkunny/go-llvm"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

func (self *LLVMOutputer) codegenConst(ir mir.Const)llvm.Constant{
	switch c := ir.(type) {
	case mir.Int:
		return self.codegenInt(c)
	case *mir.Float:
		return self.codegenFloat(c)
	case *mir.EmptyArray:
		return self.codegenEmptyArray(c)
	case *mir.EmptyStruct:
		return self.codegenEmptyStruct(c)
	case *mir.EmptyFunc:
		return self.codegenEmptyFunc(c)
	case *mir.EmptyPtr:
		return self.codegenEmptyPtr(c)
	case *mir.Array:
		return self.codegenArray(c)
	case *mir.Struct:
		return self.codegenStruct(c)
	case *mir.ConstArrayIndex:
		return self.codegenConstArrayIndex(c)
	case *mir.ConstStructIndex:
		return self.codegenConstStructIndex(c)
	case *mir.Constant:
		return self.values.Get(c).(llvm.Constant)
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenInt(ir mir.Int)llvm.Constant{
	return self.ctx.ConstInteger(self.codegenIntType(ir.Type().(mir.IntType)), ir.IntValue())
}

func (self *LLVMOutputer) codegenFloat(ir *mir.Float)llvm.Constant{
	return self.ctx.ConstFloat(self.codegenFloatType(ir.Type().(mir.FloatType)), ir.FloatValue())
}

func (self *LLVMOutputer) codegenEmptyArray(ir *mir.EmptyArray)llvm.Constant{
	return self.ctx.ConstAggregateZero(self.codegenArrayType(ir.Type().(mir.ArrayType)))
}

func (self *LLVMOutputer) codegenEmptyStruct(ir *mir.EmptyStruct)llvm.Constant{
	return self.ctx.ConstAggregateZero(self.codegenStructType(ir.Type().(mir.StructType)))
}

func (self *LLVMOutputer) codegenEmptyFunc(ir *mir.EmptyFunc)llvm.Constant{
	return self.ctx.ConstNull(self.codegenFuncTypePtr(ir.Type().(mir.FuncType)))
}

func (self *LLVMOutputer) codegenEmptyPtr(ir *mir.EmptyPtr)llvm.Constant{
	return self.ctx.ConstNull(self.codegenPtrType(ir.Type().(mir.PtrType)))
}

func (self *LLVMOutputer) codegenArray(ir *mir.Array)llvm.Constant{
	elems := lo.Map(ir.Elems(), func(item mir.Const, _ int) llvm.Constant {
		return self.codegenConst(item)
	})
	return self.ctx.ConstArray(self.codegenArrayType(ir.Type().(mir.ArrayType)), elems...)
}

func (self *LLVMOutputer) codegenStruct(ir *mir.Struct)llvm.Constant{
	elems := lo.Map(ir.Elems(), func(item mir.Const, _ int) llvm.Constant {
		return self.codegenConst(item)
	})
	return self.ctx.ConstNamedStruct(self.codegenStructType(ir.Type().(mir.StructType)), elems...)
}

func (self *LLVMOutputer) codegenConstArrayIndex(ir *mir.ConstArrayIndex)llvm.Constant{
	v := self.codegenConst(ir.Array())
	i := self.codegenConst(ir.Index())
	if ir.IsPtr(){
		return self.ctx.ConstInBoundsGEP(self.codegenType(ir.Array().Type()), v, self.ctx.ConstInteger(self.ctx.IntegerType(64), 0), i)
	}else{
		return self.ctx.ConstExtractElement(v, i)
	}
}

func (self *LLVMOutputer) codegenConstStructIndex(ir *mir.ConstStructIndex)llvm.Constant{
	v := self.codegenConst(ir.Struct())
	if ir.IsPtr(){
		return self.ctx.ConstInBoundsGEP(self.codegenType(ir.Struct().Type()), v, self.ctx.ConstInteger(self.ctx.IntegerType(32), int64(ir.Index())))
	}else{
		return self.ctx.ConstExtractElement(v, self.ctx.ConstInteger(self.ctx.IntegerType(32), int64(ir.Index())))
	}
}
