package llgen

import (
	"math/big"

	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/util"
)

func isUnitType(t hir.Type) bool {
	return stlval.Is[types.UnitType](t)
}

func (lg *LLvmGenerator) genType(t hir.Type) llvm.AnyType {
	switch t := t.(type) {
	case types.CustomType:
		if ct, ok := lg.ctx.typeCache[t.GetName()]; ok {
			if lg.ctx.definedTypes.Contain(t.GetName()) {
				return ct
			}
		}
		return lg.genCustomTypeDef(t.GetDef())
	case types.UnitType:
		return lg.ctx.emptyType
	case types.IntegerType:
		return lg.ctx.llvmCtx.Int(uint32(t.GetBits()))
	case types.FloatType:
		switch t.GetBits() {
		case 32:
			return lg.ctx.llvmCtx.Float(llvm.FloatSingle)
		case 64:
			return lg.ctx.llvmCtx.Float(llvm.FloatDouble)
		default:
			panic("unreachable")
		}
	case types.BooleanType:
		return lg.ctx.llvmCtx.Bool()
	case types.StringType:
		return lg.ctx.strType
	case types.FuncType:
		return lg.ctx.fatFuncType
	case types.TupleType:
		return lg.genTupleType(t)
	case types.ArrayType:
		return lg.genArrayType(t)
	case types.UnionType:
		return lg.genUnionType(t)
	case types.RefType:
		return lg.ctx.ptrType
	case types.StructType:
		return lg.genStructType(t)
	default:
		panic("unreachable")
	}
}

func (lg *LLvmGenerator) genCustomTypeDecl(global *globals.TypeDef) {
	switch global.Underlying.(type) {
	case types.TupleType, types.ArrayType, types.UnionType, types.FuncType, types.RefType, types.StructType:
	default:
		return
	}

	if _, ok := lg.ctx.typeCache[global.Name]; ok {
		return
	}
	name := stableName(lg.pkg, global.Name)
	lg.ctx.typeCache[global.Name] = lg.ctx.llvmCtx.NamedStruct(name)
}

func (lg *LLvmGenerator) genCustomTypeDef(global *globals.TypeDef) llvm.AnyType {
	t, _ := lg.ctx.typeCache[global.Name]
	if lg.ctx.definedTypes.Contain(global.Name) {
		return t
	}

	switch underlying := global.Underlying.(type) {
	case types.TupleType:
		llvm.MustStructType(t).SetBody(lg.genTupleType(underlying).Elems(), false)
	case types.ArrayType:
		llvm.MustStructType(t).SetBody(lg.genArrayType(underlying).Elems(), false)
	case types.UnionType:
		llvm.MustStructType(t).SetBody(lg.genUnionType(underlying).Elems(), false)
	case types.FuncType:
		llvm.MustStructType(t).SetBody(lg.ctx.fatFuncType.Elems(), false)
	case types.RefType:
		llvm.MustStructType(t).SetBody([]llvm.AnyType{lg.ctx.ptrType}, false)
	case types.StructType:
		llvm.MustStructType(t).SetBody(lg.genStructType(underlying).Elems(), false)
	default:
		lg.ctx.typeCache[global.Name] = lg.genType(global.Underlying)
	}
	lg.ctx.definedTypes.Add(global.Name)
	return lg.ctx.typeCache[global.Name]
}

// 原生函数类型
func (lg *LLvmGenerator) genNativeFuncType(t types.FuncType) llvm.FnType {
	r := lg.genFuncReturnType(t.GetReturn())
	ps := stlslices.Map(t.GetParams(), func(_ int, p hir.Type) llvm.AnyType {
		return lg.genType(p)
	})
	return lg.ctx.llvmCtx.Fn(r, ps, false)
}

// 函数返回类型，unit映射为void
func (lg *LLvmGenerator) genFuncReturnType(t hir.Type) llvm.AnyType {
	if isUnitType(t) {
		return lg.ctx.llvmCtx.Void()
	}
	return lg.genType(t)
}

func (lg *LLvmGenerator) genTupleType(t types.TupleType) llvm.StructType {
	elems := stlslices.Map(t.GetElems(), func(_ int, elem hir.Type) llvm.AnyType {
		return lg.genType(elem)
	})
	return lg.ctx.llvmCtx.Struct(elems, false)
}

func (lg *LLvmGenerator) genArrayType(t types.ArrayType) llvm.StructType {
	elem := lg.genType(t.GetElem())
	return lg.ctx.llvmCtx.Struct([]llvm.AnyType{
		lg.ctx.llvmCtx.Array(elem, typeSize(t.GetSize())),
	}, false)
}

func (lg *LLvmGenerator) genUnionType(t types.UnionType) llvm.StructType {
	key := t.String()
	if ut, ok := lg.ctx.typeCache[key]; ok {
		return llvm.MustStructType(ut)
	}

	maxSize, maxAlign := uint64(0), uint64(1)
	elems := stlslices.Map(t.GetElems(), func(_ int, elem hir.Type) llvm.AnyType {
		return lg.genType(elem)
	})
	for _, e := range elems {
		size, align := lg.typeLayout(e)
		if size > maxSize {
			maxSize = size
		}
		if align > maxAlign {
			maxAlign = align
		}
	}
	totalSize := util.AlignUp(maxSize, maxAlign)
	if maxAlign > 8 {
		panic("unreachable")
	}
	storage := lg.ctx.llvmCtx.Array(lg.ctx.i8Type, totalSize)
	if totalSize > 0 {
		storage = lg.ctx.llvmCtx.Array(lg.ctx.llvmCtx.Int(uint32(maxAlign*8)), totalSize/maxAlign)
	}
	ut := lg.ctx.llvmCtx.Struct([]llvm.AnyType{lg.ctx.i8Type, storage}, false)

	lg.ctx.typeCache[key] = ut
	return ut
}

func (lg *LLvmGenerator) genStructType(t types.StructType) llvm.StructType {
	elems := stlslices.Map(t.GetFields(), func(_ int, f *types.StructField) llvm.AnyType {
		return lg.genType(f.Type)
	})
	return lg.ctx.llvmCtx.Struct(elems, false)
}

// typeSize 类型的大小
func typeSize(size *big.Int) uint64 {
	if !size.IsUint64() {
		panic("unreachable")
	}
	return size.Uint64()
}

// typeLayout 计算类型的大小与对齐（LP64）
func (lg *LLvmGenerator) typeLayout(t llvm.AnyType) (uint64, uint64) {
	dyn := t.DynType()
	if _, err := dyn.As[llvm.IntT](); err == nil {
		bits := uint64(llvm.MustIntType(t).Bits())
		size := bits / 8
		if size == 0 {
			size = 1
		}
		return size, size
	}
	if _, err := dyn.As[llvm.FloatT](); err == nil {
		switch llvm.MustFloatType(t).Kind() {
		case llvm.FloatSingle:
			return 4, 4
		case llvm.FloatDouble:
			return 8, 8
		default:
			panic("unreachable")
		}
	}
	if _, err := dyn.As[llvm.PtrT](); err == nil {
		return 8, 8
	}
	if _, err := dyn.As[llvm.ArrayT](); err == nil {
		at := llvm.MustArrayType(t)
		size, align := lg.typeLayout(at.Elem())
		return size * at.Len(), align
	}
	if _, err := dyn.As[llvm.StructT](); err == nil {
		size, align := uint64(0), uint64(1)
		for elem := range llvm.MustStructType(t).AllElems() {
			elemSize, elemAlign := lg.typeLayout(elem)
			size = util.AlignUp(size, elemAlign) + elemSize
			if elemAlign > align {
				align = elemAlign
			}
		}
		return util.AlignUp(size, align), align
	}
	panic("unreachable")
}
