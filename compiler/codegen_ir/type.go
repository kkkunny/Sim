package codegen_ir

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/pair"
	stlos "github.com/kkkunny/stl/os"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/runtime/types"
)

func (self *CodeGenerator) codegenTypeOnly(ir hir.Type)mir.Type{
	t, _ := self.codegenType(ir)
	return t
}

func (self *CodeGenerator) codegenType(ir hir.Type) (mir.Type, types.Type) {
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

func (self *CodeGenerator) codegenEmptyType() (mir.VoidType, *types.EmptyType) {
	return self.ctx.Void(), types.TypeEmpty
}

func (self *CodeGenerator) codegenSintType(ir *hir.SintType) (mir.SintType, *types.SintType) {
	if ir.Bits == 0{
		return self.ctx.Isize(), types.TypeIsize
	}
	return self.ctx.NewSintType(stlos.Size(ir.Bits)), types.NewSintType(uint64(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *hir.UintType) (mir.UintType, *types.UintType) {
	if ir.Bits == 0{
		return self.ctx.Usize(), types.TypeUsize
	}
	return self.ctx.NewUintType(stlos.Size(ir.Bits)), types.NewUintType(uint64(ir.Bits))
}

func (self *CodeGenerator) codegenFloatType(ir *hir.FloatType) (mir.FloatType, *types.FloatType) {
	switch ir.Bits {
	case 32:
		return self.ctx.F32(), types.NewFloatType(32)
	case 64:
		return self.ctx.F64(), types.NewFloatType(64)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(ir *hir.FuncType) (mir.FuncType, *types.FuncType) {
	ret, retRt := self.codegenType(ir.Ret)
	params, paramRts := make([]mir.Type, len(ir.Params)), make([]types.Type, len(ir.Params))
	for i, p := range ir.Params{
		params[i], paramRts[i] = self.codegenType(p)
	}
	return self.ctx.NewFuncType(ret, params...), types.NewFuncType(retRt, )
}

func (self *CodeGenerator) codegenBoolType() (mir.UintType, *types.BoolType) {
	return self.ctx.NewUintType(1), types.TypeBool
}

func (self *CodeGenerator) codegenArrayType(ir *hir.ArrayType) (mir.ArrayType, *types.ArrayType) {
	elem, elemRt := self.codegenType(ir.Elem)
	return self.ctx.NewArrayType(ir.Size, elem), types.NewArrayType(uint64(ir.Size), elemRt)
}

func (self *CodeGenerator) codegenTupleType(ir *hir.TupleType) (mir.StructType, *types.TupleType) {
	elems, elemRts := make([]mir.Type, len(ir.Elems)), make([]types.Type, len(ir.Elems))
	for i, e := range ir.Elems{
		elems[i], elemRts[i] = self.codegenType(e)
	}
	return self.ctx.NewStructType(elems...), types.NewTupleType(elemRts...)
}

func (self *CodeGenerator) codegenStructType(ir *hir.StructType) (mir.StructType, *types.StructType) {
	pair := self.structs.Get(ir.String())
	return pair.First, pair.Second
}

func (self *CodeGenerator) codegenStringType() (mir.StructType, *types.StringType) {
	st, ok := self.module.NamedStructType("str")
	if ok {
		return st, types.TypeStr
	}
	return self.module.NewNamedStructType("str", self.ctx.NewPtrType(self.ctx.U8()), self.ctx.Usize()), types.TypeStr
}

func (self *CodeGenerator) codegenUnionType(ir *hir.UnionType) (mir.StructType, *types.UnionType) {
	var maxSizeType mir.Type
	var maxSize stlos.Size
	elemRts := make([]types.Type, len(ir.Elems))
	for i, e := range ir.Elems{
		var et mir.Type
		et, elemRts[i] = self.codegenType(e)
		if hir.IsUnionType(e){
			et = et.(mir.StructType).Elems()[0]
		}
		if esize := et.Size(); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	return self.ctx.NewStructType(maxSizeType, self.ctx.U8()), types.NewUnionType(elemRts...)
}

func (self *CodeGenerator) codegenPtrType(ir *hir.PtrType) (mir.PtrType, *types.PtrType) {
	elem, elemRt := self.codegenType(ir.Elem)
	return self.ctx.NewPtrType(elem), types.NewPtrType(elemRt)
}

func (self *CodeGenerator) codegenRefType(ir *hir.RefType) (mir.PtrType, *types.RefType) {
	elem, elemRt := self.codegenType(ir.Elem)
	return self.ctx.NewPtrType(elem), types.NewRefType(elemRt)
}
func (self *CodeGenerator) codegenSelfType(ir *hir.SelfType)(mir.Type, types.Type){
	return self.codegenType(ir.Self.MustValue())
}

func (self *CodeGenerator) codegenAliasType(ir *hir.AliasType)(mir.Type, types.Type){
	// TODO: 处理类型循环
	return self.codegenType(ir.Target)
}

func (self *CodeGenerator) codegenGenericIdentType(ir *hir.GenericIdentType)(mir.Type, types.Type){
	pair := self.genericIdentMapStack.Peek().Get(ir)
	return pair.First, pair.Second
}

func (self *CodeGenerator) codegenGenericStructInst(ir *hir.GenericStructInst)(mir.StructType, types.Type){
	key := fmt.Sprintf("generic_struct(%p)<%s>", ir.Define, strings.Join(stlslices.Map(ir.Params, func(i int, e hir.Type) string {
		pt, _ := self.codegenType(e)
		return pt.String()
	}), ","))
	if stp := self.structCache.Get(key); stp.First != nil{
		return stp.First, stp.Second
	}

	st, stRt := self.declGenericStructDef(ir)
	self.structCache.Set(key, pair.NewPair(st, stRt))
	self.defGenericStructDef(ir, st, stRt)
	return st, stRt
}
