package llgen

import (
	"fmt"

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
		if tt, ok := c.ctx.typeCache[t.GetDef()]; ok {
			return tt
		}
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
			panic(fmt.Errorf("llgen: 暂不支持 %d 位浮点", t.GetBits()))
		}
	case types.BooleanType:
		return c.ctx.LLVM().Bool()
	case types.StringType:
		return c.genStrType()
	case types.RefType:
		return c.ctx.LLVM().Ptr(0)
	case types.FuncType:
		return c.genFuncType(t)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的类型 %s（%T）", t, t))
	}
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

// genFuncType 函数胖类型：{ fn*, ctx* }
func (c *CodeGenerator) genFuncType(_ types.FuncType) llvm.StructType {
	return c.ctx.LLVM().Struct([]llvm.AnyType{
		c.ctx.LLVM().Ptr(0),
		c.ctx.LLVM().Ptr(0),
	}, false)
}
