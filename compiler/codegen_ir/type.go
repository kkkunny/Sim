package codegen_ir

import (
	"fmt"
	"strings"

	stlos "github.com/kkkunny/stl/os"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenType(ir hir.Type) mir.Type {
	switch t := ir.(type) {
	case *hir.EmptyType:
		return self.codegenEmptyType()
	case *hir.SintType:
		return self.codegenSintType(t)
	case *hir.UintType:
		return self.codegenUintType(t)
	case *hir.FloatType:
		return self.codegenFloatType(t)
	case *hir.FuncType:
		return self.codegenFuncType(t)
	case *hir.BoolType:
		return self.codegenBoolType()
	case *hir.ArrayType:
		return self.codegenArrayType(t)
	case *hir.TupleType:
		return self.codegenTupleType(t)
	case *hir.StructType:
		return self.codegenStructType(t)
	case *hir.StringType:
		return self.codegenStringType()
	case *hir.UnionType:
		return self.codegenUnionType(t)
	case *hir.PtrType:
		return self.codegenPtrType(t)
	case *hir.RefType:
		return self.codegenRefType(t)
	case *hir.SelfType:
		return self.codegenSelfType(t)
	case *hir.AliasType:
		return self.codegenAliasType(t)
	case *hir.GenericIdentType:
		return self.codegenGenericIdentType(t)
	case *hir.GenericStructInst:
		return self.codegenGenericStructInst(t)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType() mir.VoidType {
	return self.ctx.Void()
}

func (self *CodeGenerator) codegenSintType(ir *hir.SintType) mir.SintType {
	if ir.Bits == 0{
		return self.ctx.Isize()
	}
	return self.ctx.NewSintType(stlos.Size(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *hir.UintType) mir.UintType {
	if ir.Bits == 0{
		return self.ctx.Usize()
	}
	return self.ctx.NewUintType(stlos.Size(ir.Bits))
}

func (self *CodeGenerator) codegenFloatType(ir *hir.FloatType) mir.FloatType {
	switch ir.Bits {
	case 32:
		return self.ctx.F32()
	case 64:
		return self.ctx.F64()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(ir *hir.FuncType) mir.FuncType {
	ret := self.codegenType(ir.Ret)
	params := lo.Map(ir.Params, func(item hir.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewFuncType(ret, params...)
}

func (self *CodeGenerator) codegenBoolType() mir.UintType {
	return self.ctx.NewUintType(1)
}

func (self *CodeGenerator) codegenArrayType(ir *hir.ArrayType) mir.ArrayType {
	return self.ctx.NewArrayType(ir.Size, self.codegenType(ir.Elem))
}

func (self *CodeGenerator) codegenTupleType(ir *hir.TupleType) mir.StructType {
	elems := lo.Map(ir.Elems, func(item hir.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewStructType(elems...)
}

func (self *CodeGenerator) codegenStructType(ir *hir.StructType) mir.StructType {
	return self.structs.Get(ir)
}

func (self *CodeGenerator) codegenStringType() mir.StructType {
	st, ok := self.module.NamedStructType("str")
	if ok {
		return st
	}
	return self.module.NewNamedStructType("str", self.ctx.NewPtrType(self.ctx.U8()), self.ctx.Usize())
}

func (self *CodeGenerator) codegenUnionType(ir *hir.UnionType) mir.StructType {
	var maxSizeType mir.Type
	var maxSize stlos.Size
	for _, e := range ir.Elems{
		et := self.codegenType(e)
		if esize := et.Size(); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	return self.ctx.NewStructType(maxSizeType, self.ctx.U8())
}

func (self *CodeGenerator) codegenPtrType(ir *hir.PtrType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(ir.Elem))
}

func (self *CodeGenerator) codegenRefType(ir *hir.RefType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(ir.Elem))
}
func (self *CodeGenerator) codegenSelfType(ir *hir.SelfType)mir.Type{
	return self.codegenType(ir.Self)
}

func (self *CodeGenerator) codegenAliasType(ir *hir.AliasType)mir.Type{
	// TODO: 处理类型循环
	return self.codegenType(ir.Target)
}

func (self *CodeGenerator) codegenGenericIdentType(ir *hir.GenericIdentType)mir.Type{
	return self.codegenType(self.genericIdentMapStack.Peek().Get(ir))
}

func (self *CodeGenerator) codegenGenericStructInst(ir *hir.GenericStructInst)mir.StructType{
	key := fmt.Sprintf("generic_struct(%p)<%s>", ir.Define, strings.Join(stlslices.Map(ir.Params, func(i int, e hir.Type) string {
		return self.codegenType(e).String()
	}), ","))
	if st := self.structCache.Get(key); st != nil{
		return st
	}

	st := self.declGenericStructDef(ir)
	self.structCache.Set(key, st)
	self.defGenericStructDef(ir, st)
	return st
}
