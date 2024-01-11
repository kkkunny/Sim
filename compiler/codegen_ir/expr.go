package codegen_ir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/runtime/traits"
	"github.com/kkkunny/Sim/runtime/types"
)

func (self *CodeGenerator) codegenExpr(ir hir.Expr, load bool) mir.Value {
	switch expr := ir.(type) {
	case *hir.Integer:
		return self.codegenInteger(expr)
	case *hir.Float:
		return self.codegenFloat(expr)
	case *hir.Boolean:
		return self.codegenBool(expr)
	case *hir.Assign:
		self.codegenAssign(expr)
		return nil
	case hir.Binary:
		return self.codegenBinary(expr)
	case hir.Unary:
		return self.codegenUnary(expr, load)
	case hir.Ident:
		return self.codegenIdent(expr, load)
	case *hir.Call:
		return self.codegenCall(expr)
	case hir.TypeCovert:
		return self.codegenCovert(expr, load)
	case *hir.Array:
		return self.codegenArray(expr)
	case *hir.Index:
		return self.codegenIndex(expr, load)
	case *hir.Tuple:
		return self.codegenTuple(expr)
	case *hir.Extract:
		return self.codegenExtract(expr, load)
	case *hir.Default:
		return self.codegenDefault(expr.GetType())
	case *hir.Struct:
		return self.codegenStruct(expr)
	case *hir.GetField:
		return self.codegenField(expr, load)
	case *hir.String:
		return self.codegenString(expr)
	case *hir.Union:
		return self.codegenUnion(expr, load)
	case *hir.TypeJudgment:
		return self.codegenTypeJudgment(expr)
	case *hir.UnUnion:
		return self.codegenUnUnion(expr)
	case *hir.WrapWithNull:
		return self.codegenWrapWithNull(expr, load)
	case *hir.CheckNull:
		return self.codegenCheckNull(expr)
	case *hir.MethodDef, *hir.GenericStructMethodInst, *hir.GenericMethodInst, *hir.GenericStructGenericMethodInst:
		// TODO: 闭包
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(ir *hir.Integer) mir.Int {
	return mir.NewInt(self.codegenTypeOnly(ir.Type).(mir.IntType), ir.Value.Int64())
}

func (self *CodeGenerator) codegenFloat(ir *hir.Float) *mir.Float {
	v, _ := ir.Value.Float64()
	return mir.NewFloat(self.codegenTypeOnly(ir.Type).(mir.FloatType), v)
}

func (self *CodeGenerator) codegenBool(ir *hir.Boolean) *mir.Uint {
	return mir.Bool(self.ctx, ir.Value)
}

func (self *CodeGenerator) codegenAssign(ir *hir.Assign) {
	if l, ok := ir.Left.(*hir.Tuple); ok {
		self.codegenUnTuple(ir.Right, l.Elems)
	} else {
		left, right := self.codegenExpr(ir.GetLeft(), false), self.codegenExpr(ir.GetRight(), true)
		self.builder.BuildStore(right, left)
	}
}

func (self *CodeGenerator) codegenUnTuple(fromIr hir.Expr, toIrs []hir.Expr) {
	if tupleNode, ok := fromIr.(*hir.Tuple); ok {
		for i, l := range toIrs {
			self.codegenAssign(&hir.Assign{
				Left:  l,
				Right: tupleNode.Elems[i],
			})
		}
	} else {
		var unTuple func(from mir.Value, toNodes []hir.Expr)
		unTuple = func(from mir.Value, toNodes []hir.Expr) {
			for i, toNode := range toNodes {
				if toNodes, ok := toNode.(*hir.Tuple); ok {
					index := self.buildStructIndex(from, uint64(i))
					unTuple(index, toNodes.Elems)
				} else {
					value := self.buildStructIndex(from, uint64(i), false)
					to := self.codegenExpr(toNode, false)
					self.builder.BuildStore(value, to)
				}
			}
		}
		from := self.codegenExpr(fromIr, false)
		unTuple(from, toIrs)
	}
}

func (self *CodeGenerator) codegenBinary(ir hir.Binary) mir.Value {
	left, right := self.codegenExpr(ir.GetLeft(), true), self.codegenExpr(ir.GetRight(), true)
	switch ir.(type) {
	case *hir.IntAndInt, *hir.BoolAndBool:
		return self.builder.BuildAnd(left, right)
	case *hir.IntOrInt, *hir.BoolOrBool:
		return self.builder.BuildOr(left, right)
	case *hir.IntXorInt:
		return self.builder.BuildXor(left, right)
	case *hir.IntShlInt:
		return self.builder.BuildShl(left, right)
	case *hir.IntShrInt:
		return self.builder.BuildShr(left, right)
	case *hir.NumAddNum:
		return self.builder.BuildAdd(left, right)
	case *hir.NumSubNum:
		return self.builder.BuildSub(left, right)
	case *hir.NumMulNum:
		return self.builder.BuildMul(left, right)
	case *hir.NumDivNum:
		self.buildCheckZero(right)
		return self.builder.BuildDiv(left, right)
	case *hir.NumRemNum:
		self.buildCheckZero(right)
		return self.builder.BuildRem(left, right)
	case *hir.NumLtNum:
		return self.builder.BuildCmp(mir.CmpKindLT, left, right)
	case *hir.NumGtNum:
		return self.builder.BuildCmp(mir.CmpKindGT, left, right)
	case *hir.NumLeNum:
		return self.builder.BuildCmp(mir.CmpKindLE, left, right)
	case *hir.NumGeNum:
		return self.builder.BuildCmp(mir.CmpKindGE, left, right)
	case *hir.Equal:
		return self.buildEqual(ir.GetLeft().GetType(), left, right, false)
	case *hir.NotEqual:
		return self.buildEqual(ir.GetLeft().GetType(), left, right, true)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(ir hir.Unary, load bool) mir.Value {
	switch ir.(type) {
	case *hir.NumNegate:
		return self.builder.BuildNeg(self.codegenExpr(ir.GetValue(), true))
	case *hir.IntBitNegate, *hir.BoolNegate:
		return self.builder.BuildNot(self.codegenExpr(ir.GetValue(), true))
	case *hir.GetPtr:
		return self.codegenExpr(ir.GetValue(), false)
	case *hir.GetValue:
		ptr := self.codegenExpr(ir.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(ir hir.Ident, load bool) mir.Value {
	switch identNode := ir.(type) {
	case *hir.FuncDef:
		return self.values.Get(identNode).(*mir.Function)
	case hir.Variable:
		p := self.values.Get(identNode)
		if !load {
			return p
		}
		return self.builder.BuildLoad(p)
	case *hir.GenericFuncInst:
		return self.codegenGenericFuncInst(identNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(ir *hir.Call) mir.Value {
	args := lo.Map(ir.Args, func(item hir.Expr, index int) mir.Value {
		return self.codegenExpr(item, true)
	})
	var f, selfValue mir.Value
	switch fnIr := ir.Func.(type) {
	case hir.MethodExpr:
		if selfValueIr, ok := fnIr.GetSelf(); ok {
			selfValue = self.codegenExpr(selfValueIr, false)
		}
		switch method := fnIr.(type) {
		case *hir.Method:
			f = self.values.Get(&method.Define.FuncDef)
		case *hir.GenericStructMethodInst:
			f = self.codegenGenericStructMethodInst(method)
		case *hir.GenericMethodInst:
			f = self.codegenGenericMethodInst(method)
		case *hir.GenericStructGenericMethodInst:
			f = self.codegenGenericStructGenericMethodInst(method)
		default:
			panic("unreachable")
		}
	default:
		f = self.codegenExpr(ir.Func, true)
	}
	if selfValue != nil {
		var selfRef mir.Value
		if expectSelfRefType := f.Type().(mir.FuncType).Params()[0]; !selfValue.Type().Equal(expectSelfRefType){
			selfRef = self.builder.BuildAllocFromStack(selfValue.Type())
			self.builder.BuildStore(selfValue, selfRef)
		}else{
			selfRef = selfValue
		}
		args = append([]mir.Value{selfRef}, args...)
	}
	return self.builder.BuildCall(f, args...)
}

func (self *CodeGenerator) codegenCovert(ir hir.TypeCovert, load bool) mir.Value {
	switch ir.(type) {
	case *hir.Num2Num:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenTypeOnly(ir.GetType())
		return self.builder.BuildNumberCovert(from, to.(mir.NumberType))
	case *hir.Pointer2Pointer:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenTypeOnly(ir.GetType())
		return self.builder.BuildPtrToPtr(from, to.(mir.PtrType))
	case *hir.Pointer2Usize:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenTypeOnly(ir.GetType())
		return self.builder.BuildPtrToUint(from, to.(mir.UintType))
	case *hir.Usize2Pointer:
		from := self.codegenExpr(ir.GetFrom(), true)
		to := self.codegenTypeOnly(ir.GetType())
		return self.builder.BuildUintToPtr(from, to.(mir.GenericPtrType))
	case *hir.ShrinkUnion:
		_, srcRt := self.codegenType(ir.GetFrom().GetType())
		dst, dstRt := self.codegenType(ir.GetType())
		from := self.codegenExpr(ir.GetFrom(), false)
		srcData := self.buildStructIndex(from, 0, false)
		srcIndex := self.buildStructIndex(from, 1, false)
		newIndex := self.buildCovertUnionIndex(srcRt.(*types.UnionType), dstRt.(*types.UnionType), srcIndex)
		ptr := self.builder.BuildAllocFromStack(dst)
		newDataPtr := self.builder.BuildPtrToPtr(self.buildStructIndex(ptr, 0, true), self.ctx.NewPtrType(srcData.Type()))
		self.builder.BuildStore(srcData, newDataPtr)
		newIndexPtr := self.buildStructIndex(ptr, 1, true)
		self.builder.BuildStore(newIndex, newIndexPtr)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	case *hir.ExpandUnion:
		_, srcRt := self.codegenType(ir.GetFrom().GetType())
		dst, dstRt := self.codegenType(ir.GetType())
		from := self.codegenExpr(ir.GetFrom(), false)
		srcData := self.buildStructIndex(from, 0, false)
		srcIndex := self.buildStructIndex(from, 1, false)
		newIndex := self.buildCovertUnionIndex(srcRt.(*types.UnionType), dstRt.(*types.UnionType), srcIndex)
		ptr := self.builder.BuildAllocFromStack(dst)
		newDataPtr := self.builder.BuildPtrToPtr(self.buildStructIndex(ptr, 0, true), self.ctx.NewPtrType(srcData.Type()))
		self.builder.BuildStore(srcData, newDataPtr)
		newIndexPtr := self.buildStructIndex(ptr, 1, true)
		self.builder.BuildStore(newIndex, newIndexPtr)
		if !load {
			return ptr
		}
		return self.builder.BuildLoad(ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(ir *hir.Array) mir.Value {
	elems := lo.Map(ir.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackArray(self.codegenTypeOnly(ir.GetType()).(mir.ArrayType), elems...)
}

func (self *CodeGenerator) codegenIndex(ir *hir.Index, load bool) mir.Value {
	at := self.codegenTypeOnly(ir.From.GetType()).(mir.ArrayType)
	index := self.codegenExpr(ir.Index, true)
	self.buildCheckIndex(index, uint64(at.Length()))
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildArrayIndex(from, index)
	if load && (stlbasic.Is[*mir.ArrayIndex](ptr) && ptr.(*mir.ArrayIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTuple(ir *hir.Tuple) mir.Value {
	elems := lo.Map(ir.Elems, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenTypeOnly(ir.GetType()).(mir.StructType), elems...)
}

func (self *CodeGenerator) codegenExtract(ir *hir.Extract, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenDefault(tir hir.Type) mir.Value {
	t, rtt := self.codegenType(tir)

	switch trt := rtt.(type) {
	case *types.EmptyType, *types.FuncType, *types.RefType:
		panic("unreachable")
	case *types.SintType, *types.UintType, *types.FloatType, *types.BoolType, *types.PtrType, *types.UnionType:
		return mir.NewZero(t)
	case *types.StringType:
		return self.constString("")
	case *types.ArrayType:
		// TODO: 填充所有字段为默认值
		return mir.NewZero(t)
	case *types.TupleType:
		elems := stlslices.Map(hir.AsTupleType(tir).Elems, func(_ int, e hir.Type) mir.Value {
			return self.codegenDefault(e)
		})
		return self.builder.BuildPackStruct(t.(mir.StructType), elems...)
	case *types.StructType:
		defaultTrait := traits.NewDefault(trt)
		if !defaultTrait.IsInst(trt){
			// TODO: 填充所有字段为默认值
			return mir.NewZero(t)
		}
		st := hir.AsStructType(tir)
		methodObj := st.Methods.Get(defaultTrait.Methods.Keys().Get(0))
		switch method := methodObj.(type) {
		case *hir.MethodDef:
			return self.codegenCall(&hir.Call{
				Func: &hir.Method{
					Define: method,
					Self: either.Right[hir.Expr, *hir.StructType](st),
				},
			})
		case *hir.GenericStructMethodDef:
			return self.codegenCall(&hir.Call{
				Func: &hir.GenericStructMethodInst{
					Define: method,
					Self: either.Right[hir.Expr, *hir.StructType](st),
				},
			})
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenStruct(ir *hir.Struct) mir.Value {
	fields := lo.Map(ir.Fields, func(item hir.Expr, _ int) mir.Value {
		return self.codegenExpr(item, true)
	})
	return self.builder.BuildPackStruct(self.codegenTypeOnly(ir.Type).(mir.StructType), fields...)
}

func (self *CodeGenerator) codegenField(ir *hir.GetField, load bool) mir.Value {
	from := self.codegenExpr(ir.From, false)
	ptr := self.buildStructIndex(from, uint64(ir.Index))
	if load && (stlbasic.Is[*mir.StructIndex](ptr) && ptr.(*mir.StructIndex).IsPtr()) {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenString(ir *hir.String) mir.Value {
	return self.constString(ir.Value)
}

func (self *CodeGenerator) codegenUnion(ir *hir.Union, load bool) mir.Value {
	utObj, utRtObj := self.codegenType(ir.Type)
	ut, utRt := utObj.(mir.StructType), utRtObj.(*types.UnionType)
	_, vtRtObj := self.codegenType(ir.Value.GetType())

	value := self.codegenExpr(ir.Value, true)
	ptr := self.builder.BuildAllocFromStack(ut)
	dataPtr := self.buildStructIndex(ptr, 0, true)
	dataPtr = self.builder.BuildPtrToPtr(dataPtr, self.ctx.NewPtrType(value.Type()))
	self.builder.BuildStore(value, dataPtr)

	index := utRt.IndexElem(vtRtObj)
	self.builder.BuildStore(
		mir.NewInt(ut.Elems()[1].(mir.UintType), int64(index)),
		self.buildStructIndex(ptr, 1, true),
	)

	if load {
		return self.builder.BuildLoad(ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenTypeJudgment(ir *hir.TypeJudgment) mir.Value {
	_, dstTRtObj := self.codegenType(ir.Type)
	srcTObj, srcTRtObj := self.codegenType(ir.Value.GetType())
	switch {
	case stlbasic.Is[*types.UnionType](srcTRtObj) && srcTRtObj.(*types.UnionType).Contain(dstTRtObj):
		srcT, srcRt := srcTObj.(mir.StructType), srcTRtObj.(*types.UnionType)

		from := self.codegenExpr(ir.Value, false)
		srcIndex := self.buildStructIndex(from, 1, false)

		if dstRt, ok := dstTRtObj.(*types.UnionType); ok {
			return self.buildCheckUnionType(srcRt, dstRt, srcIndex)
		} else {
			targetIndex := srcRt.IndexElem(dstTRtObj)
			return self.builder.BuildCmp(mir.CmpKindEQ, srcIndex, mir.NewInt(srcT.Elems()[1].(mir.IntType), int64(targetIndex)))
		}
	default:
		return mir.Bool(self.ctx, srcTRtObj.Equal(dstTRtObj))
	}
}

func (self *CodeGenerator) codegenUnUnion(ir *hir.UnUnion) mir.Value {
	value := self.codegenExpr(ir.Value, false)
	elemPtr := self.buildStructIndex(value, 0, true)
	elemPtr = self.builder.BuildPtrToPtr(elemPtr, self.ctx.NewPtrType(self.codegenTypeOnly(ir.GetType())))
	return self.builder.BuildLoad(elemPtr)
}

func (self *CodeGenerator) codegenWrapWithNull(ir *hir.WrapWithNull, load bool) mir.Value {
	return self.codegenExpr(ir.Value, load)
}

func (self *CodeGenerator) codegenCheckNull(ir *hir.CheckNull) mir.Value {
	ptr := self.codegenExpr(ir.Value, true)
	self.buildCheckNull(ptr)
	return ptr
}

func (self *CodeGenerator) codegenGenericFuncInst(ir *hir.GenericFuncInst) mir.Value {
	cur := self.builder.Current()
	defer func() {
		self.builder.MoveTo(cur)
	}()

	key := fmt.Sprintf("generic_func(%p)<%s>", ir.Define, strings.Join(stlslices.Map(ir.Args, func(i int, e hir.Type) string {
		return self.codegenTypeOnly(e).String()
	}), ","))
	if f := self.funcCache.Get(key); f != nil {
		return f
	}

	f := self.declGenericFuncDef(ir)
	self.funcCache.Set(key, f)
	self.defGenericFuncDef(ir, f)
	return f
}

func (self *CodeGenerator) codegenGenericStructMethodInst(ir *hir.GenericStructMethodInst) *mir.Function {
	cur := self.builder.Current()
	defer func() {
		self.builder.MoveTo(cur)
	}()

	key := fmt.Sprintf("(%p<%s>)generic_method(%p)", ir.Define.Scope, strings.Join(stlslices.Map(ir.GetGenericArgs(), func(i int, e hir.Type) string {
		return self.codegenTypeOnly(e).String()
	}), ","), ir.Define)
	if f := self.funcCache.Get(key); f != nil {
		return f
	}

	f := self.declGenericStructMethodDef(ir)
	self.funcCache.Set(key, f)
	self.defGenericStructMethodDef(ir, f)
	return f
}

func (self *CodeGenerator) codegenGenericMethodInst(ir *hir.GenericMethodInst) *mir.Function {
	cur := self.builder.Current()
	defer func() {
		self.builder.MoveTo(cur)
	}()

	key := fmt.Sprintf("(%p)generic_method(%p)<%s>", ir.Define.Scope, ir.Define, strings.Join(stlslices.Map(ir.Args, func(i int, e hir.Type) string {
		return self.codegenTypeOnly(e).String()
	}), ","))
	if f := self.funcCache.Get(key); f != nil {
		return f
	}

	f := self.declGenericMethodDef(ir)
	self.funcCache.Set(key, f)
	self.defGenericMethodDef(ir, f)
	return f
}

func (self *CodeGenerator) codegenGenericStructGenericMethodInst(ir *hir.GenericStructGenericMethodInst) *mir.Function {
	cur := self.builder.Current()
	defer func() {
		self.builder.MoveTo(cur)
	}()

	key := fmt.Sprintf("(%p<%s>)generic_method(%p)<%s>", ir.Define.Scope, strings.Join(stlslices.Map(ir.GetScopeGenericArgs(), func(i int, e hir.Type) string {
		return self.codegenTypeOnly(e).String()
	}), ","), ir.Define, strings.Join(stlslices.Map(ir.Args, func(i int, e hir.Type) string {
		return self.codegenTypeOnly(e).String()
	}), ","))
	if f := self.funcCache.Get(key); f != nil {
		return f
	}

	f := self.declGenericStructGenericMethodDef(ir)
	self.funcCache.Set(key, f)
	self.defGenericStructGenericMethodDef(ir, f)
	return f
}
