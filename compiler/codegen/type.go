package codegen

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

const strTypeName = "str"

// genType HIR 类型 → LLVM 类型
func (c *CodeGenerator) genType(t hir.Type) llvm.AnyType {
	switch t := t.(type) {
	case types.CustomType:
		// 无条件走 genCustomTypeDef：它自带 cache-hit 分支并在 opaque 时按需填充 body。
		// 若这里命中缓存直接返回，后置声明的聚合出现在 array/union 成员（填充期需要尺寸）时
		// 会因 opaque 无法计算布局而 panic（§12，M4 修复）。
		return c.genCustomTypeDef(t.GetDef())
	case types.UnitType:
		return c.ctx.LLVM().Void()
	case types.SintType:
		return c.ctx.LLVM().Int(uint32(t.GetBits()))
	case types.UintType:
		return c.ctx.LLVM().Int(uint32(t.GetBits()))
	case types.FloatType:
		switch t.GetBits() {
		case 32:
			return c.ctx.LLVM().Float(llvm.FloatSingle)
		case 64:
			return c.ctx.LLVM().Float(llvm.FloatDouble)
		default:
			panic(c.ice("%d-bit float is not supported", t.GetBits()))
		}
	case types.BooleanType:
		return c.ctx.LLVM().Bool()
	case types.StringType:
		return c.genStrType()
	case types.RefType:
		return c.ctx.LLVM().Ptr(0)
	case types.FuncType:
		return c.genFuncType(t)
	case types.TupleType:
		return c.genTupleType(t)
	case types.ArrayType:
		return c.genArrayType(t)
	case types.StructType:
		return c.genStructType(t)
	case types.UnionType:
		return c.genUnionType(t)
	default:
		panic(c.ice("unsupported type %s (%T)", t, t))
	}
}

// genTupleType 元组类型（B4）：字面量结构体，字段顺序即元素顺序
func (c *CodeGenerator) genTupleType(t types.TupleType) llvm.StructType {
	return c.ctx.LLVM().Struct(c.genTypes(t.GetElems()), false)
}

// genArrayType 数组类型（B5）：[N x T]；零尺寸（N=0 或元素零尺寸）→ {}（B10）
func (c *CodeGenerator) genArrayType(t types.ArrayType) llvm.AnyType {
	elem := c.genType(t.GetElem())
	if t.GetSize().Sign() <= 0 || c.llvmSizeOf(elem) == 0 {
		return c.genEmptyStruct()
	}
	return c.ctx.LLVM().Array(elem, t.GetSize().Uint64())
}

// genStructType 结构体类型（B6）：字面量结构体，字段顺序按 GetFields()
func (c *CodeGenerator) genStructType(t types.StructType) llvm.StructType {
	return c.ctx.LLVM().Struct(c.genTypes(stlslices.Map(t.GetFields(), func(_ int, f *types.StructField) hir.Type {
		return f.Type
	})), false)
}

// genUnionType union 类型（B9）：{ i8 tag, payload }；全零尺寸 union → {}（B10）
func (c *CodeGenerator) genUnionType(t types.UnionType) llvm.StructType {
	elems := c.unionMemberTypes(t)
	if !unionHasPayload(elems) {
		return c.genEmptyStruct()
	}
	payload, _ := c.genUnionPayload(elems)
	return c.ctx.LLVM().Struct([]llvm.AnyType{c.ctx.LLVM().Int(8), payload}, false)
}

// genTypes HIR 类型列表 → LLVM 类型列表
func (c *CodeGenerator) genTypes(ts []hir.Type) []llvm.AnyType {
	return stlslices.Map(ts, func(_ int, t hir.Type) llvm.AnyType {
		return c.genType(t)
	})
}

// genEmptyStruct 空结构体 {}（零尺寸类型 B10）
func (c *CodeGenerator) genEmptyStruct() llvm.StructType {
	return c.ctx.LLVM().Struct(nil, false)
}

// genZeroValue 任意类型的零值常量（聚合类型得到 zeroinitializer）
func (c *CodeGenerator) genZeroValue(t llvm.AnyType) llvm.AnyValue {
	return c.ctx.LLVM().ConstZero(t.DynType())
}

// llvmSizeOf 类型的 ABI 尺寸（byte）；void 视为 0
func (c *CodeGenerator) llvmSizeOf(t llvm.AnyType) uint64 {
	if _, ok := t.(llvm.VoidType); ok {
		return 0
	}
	if !t.IsSized() {
		panic(c.ice("type %s has no size (opaque), cannot compute layout", t))
	}
	return c.ctx.DataLayout().ABISizeOfType(t)
}

// isZeroSizeLLVM LLVM 类型是否零尺寸（含 void）
func (c *CodeGenerator) isZeroSizeLLVM(t llvm.AnyType) bool {
	return c.llvmSizeOf(t) == 0
}

// unionMemberTypes union 各成员的 LLVM 类型，下标与 HIR 成员一一对应；
// 零尺寸成员（含 unit）为 nil，表示不占载荷。
func (c *CodeGenerator) unionMemberTypes(t types.UnionType) []llvm.AnyType {
	elems := make([]llvm.AnyType, len(t.GetElems()))
	for i, e := range t.GetElems() {
		if et := c.genType(e); c.llvmSizeOf(et) > 0 {
			elems[i] = et
		}
	}
	return elems
}

// unionHasPayload 是否存在非零尺寸成员
func unionHasPayload(elems []llvm.AnyType) bool {
	return stlslices.Any(elems, func(_ int, t llvm.AnyType) bool { return t != nil })
}

// genUnionPayload 计算 union payload（§4.5）：取对齐要求最大的成员 m，
// payload = struct { m, [size(union)-size(m) x i8] }，其中
// size(union) = roundup(max成员size, max对齐)，用 DataLayout 精确计算。
// 返回 payload 类型与选中成员下标；无任何非零尺寸成员时返回 {} 与 -1。
func (c *CodeGenerator) genUnionPayload(elems []llvm.AnyType) (llvm.StructType, int) {
	dl := c.ctx.DataLayout()
	var member int = -1
	var maxAlign, maxSize uint64
	for i, et := range elems {
		if et == nil {
			continue
		}
		align := uint64(dl.ABIAlignOfType(et))
		size := dl.ABISizeOfType(et)
		// 对齐要求最大的成员作为载荷首字段（同对齐取尺寸更大者，减少填充）
		if member < 0 || align > maxAlign || (align == maxAlign && size > dl.ABISizeOfType(elems[member])) {
			member, maxAlign = i, align
		}
		if size > maxSize {
			maxSize = size
		}
	}
	if member < 0 {
		return c.genEmptyStruct(), -1
	}
	// size(union) = roundup(最大成员 size, 最大对齐)
	unionSize := (maxSize + maxAlign - 1) / maxAlign * maxAlign
	payloadElems := []llvm.AnyType{elems[member]}
	if pad := unionSize - dl.ABISizeOfType(elems[member]); pad > 0 {
		payloadElems = append(payloadElems, c.ctx.LLVM().Array(c.ctx.LLVM().Int(8), pad))
	}
	return c.ctx.LLVM().Struct(payloadElems, false), member
}

// genStrType str 类型：{ const char* data; i64 length; }
func (c *CodeGenerator) genStrType() llvm.StructType {
	return c.ctx.NamedStructWithBody(strTypeName,
		c.ctx.LLVM().Ptr(0),
		c.ctx.LLVM().Int(64),
	)
}

// genNativeFuncType 原生函数类型（无闭包 ctx）
func (c *CodeGenerator) genNativeFuncType(t types.FuncType) llvm.FnType {
	r := c.genType(t.GetReturn())
	ps := stlslices.Map(t.GetParams(), func(_ int, e hir.Type) llvm.AnyType {
		return c.genType(e)
	})
	return c.ctx.LLVM().Fn(r, ps, false)
}

// genFuncType 函数胖类型：{ fn*, ctx* }（B8/§4.4）
func (c *CodeGenerator) genFuncType(_ types.FuncType) llvm.StructType {
	return c.genFatFuncType()
}

// genFatFuncType 胖函数值的 LLVM 表示 { ptr fn, ptr ctx }；
// 所有函数值共用同一字面量结构（LLVM 结构类型按结构相等）
func (c *CodeGenerator) genFatFuncType() llvm.StructType {
	return c.ctx.LLVM().Struct([]llvm.AnyType{
		c.ctx.LLVM().Ptr(0),
		c.ctx.LLVM().Ptr(0),
	}, false)
}

// genCtxFuncType 带闭包 ctx 的调用签名：首参 ptr ctx，其余同原生签名（F1/F4）
func (c *CodeGenerator) genCtxFuncType(t types.FuncType) llvm.FnType {
	return c.ctx.LLVM().Fn(c.genType(t.GetReturn()),
		append([]llvm.AnyType{c.ctx.LLVM().Ptr(0)}, c.genTypes(t.GetParams())...), false)
}
