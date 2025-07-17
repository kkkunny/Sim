package codegen_ir

import (
	"slices"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

func (self *CodeGenerator) codegenValue(ir hir.Value, load bool) llvm.Value {
	switch ir := ir.(type) {
	case *values.Integer:
		return self.codegenInteger(ir)
	case *values.Float:
		return self.codegenFloat(ir)
	case *values.Boolean:
		return self.builder.ConstBoolean(ir.Value())
	case *values.String:
		return self.builder.ConstString(ir.Value())
	case local.BinaryExpr:
		return self.codegenBinary(ir, load)
	case local.UnaryExpr:
		return self.codegenUnary(ir, load)
	case values.Ident:
		return self.codegenIdent(ir, load)
	case *local.CallExpr:
		return self.codegenCall(ir)
	case local.CovertExpr:
		return self.codegenCovert(ir, load)
	case *local.ArrayExpr:
		return self.codegenArray(ir)
	case *local.TupleExpr:
		return self.codegenTuple(ir)
	case *local.DefaultExpr:
		return self.codegenDefault(ir.Type())
	case *local.StructExpr:
		return self.codegenStruct(ir)
	case *local.TypeJudgmentExpr:
		return self.codegenTypeJudgment(ir)
	case *local.LambdaExpr:
		return self.codegenLambda(ir)
	case *local.EnumExpr:
		return self.codegenEnum(ir)
	case *local.GenericFuncInstExpr:
		return self.codegenGenericFuncInst(ir)
	case *local.StaticMethodExpr:
		return self.codegenStaticMethod(ir)
	case *local.TraitStaticMethodExpr:
		return self.codegenTraitStaticMethod(ir)
	case *local.TraitMethodExpr:
		return self.codegenValue(self.covertTraitMethodToMethod(ir), load)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(ir *values.Integer) llvm.ConstInteger {
	return self.builder.ConstInteger(self.codegenType(ir.Type()).(llvm.IntegerType), ir.Value().Int64())
}

func (self *CodeGenerator) codegenFloat(ir *values.Float) llvm.ConstFloat {
	return self.builder.ConstFloat(self.codegenType(ir.Type()).(llvm.FloatType), stlval.IgnoreWith(ir.Value().Float64()))
}

func (self *CodeGenerator) codegenAssign(ir *local.AssignExpr) {
	if tIr, ok := ir.GetLeft().(*local.TupleExpr); ok {
		self.codegenUnTuple(ir.GetRight(), tIr.Elems())
	} else {
		left, right := self.codegenValue(ir.GetLeft(), false), self.codegenValue(ir.GetRight(), true)
		self.builder.CreateStore(right, left)
	}
}

func (self *CodeGenerator) codegenUnTuple(fromIr hir.Value, toIrs []hir.Value) {
	if fromTIr, ok := fromIr.(*local.TupleExpr); ok {
		for i, toIr := range toIrs {
			self.codegenAssign(local.NewAssignExpr(toIr, fromTIr.Elems()[i]))
		}
	} else {
		var unTuple func(tIr hir.Type, from llvm.Value, toIrs []hir.Value)
		unTuple = func(ftIr hir.Type, from llvm.Value, toIrs []hir.Value) {
			fttIr := stlval.IgnoreWith(types.As[types.TupleType](ftIr))
			ftt := self.codegenTupleType(fttIr)
			for i, toIr := range toIrs {
				if toTIr, ok := toIr.(*local.TupleExpr); ok {
					index := self.builder.CreateStructIndex(ftt, from, uint(i))
					unTuple(fttIr.Elems()[i], index, toTIr.Elems())
				} else {
					value := self.builder.CreateStructIndex(ftt, from, uint(i), false)
					to := self.codegenValue(toIr, false)
					self.builder.CreateStore(self.buildCopy(fttIr.Elems()[i], value), to)
				}
			}
		}
		from := self.codegenValue(fromIr, false)
		unTuple(fromIr.Type(), from, toIrs)
	}
}

func (self *CodeGenerator) codegenBinary(ir local.BinaryExpr, load bool) llvm.Value {
	lt := ir.GetLeft().Type()
	switch ir := ir.(type) {
	case *local.AssignExpr:
		self.codegenAssign(ir)
		return nil
	case *local.AndExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("And")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateAnd("", left, right)
	case *local.LogicAndExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Land")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateAnd("", left, right)
	case *local.OrExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Or")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateOr("", left, right)
	case *local.LogicOrExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lor")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateOr("", left, right)
	case *local.XorExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Xor")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateXor("", left, right)
	case *local.ShlExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Shl")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateShl("", left, right)
	case *local.ShrExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Shr")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.SintType](lt) {
			return self.builder.CreateAShr("", left, right)
		} else {
			return self.builder.CreateLShr("", left, right)
		}
	case *local.AddExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Add")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFAdd("", left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateSAdd("", left, right)
		} else {
			return self.builder.CreateUAdd("", left, right)
		}
	case *local.SubExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Sub")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFSub("", left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateSSub("", left, right)
		} else {
			return self.builder.CreateUSub("", left, right)
		}
	case *local.MulExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Mul")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFMul("", left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateSMul("", left, right)
		} else {
			return self.builder.CreateUMul("", left, right)
		}
	case *local.DivExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Div")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		self.buildCheckZero(right)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFDiv("", left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateSDiv("", left, right)
		} else {
			return self.builder.CreateUDiv("", left, right)
		}
	case *local.RemExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Rem")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		self.buildCheckZero(right)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFRem("", left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateSRem("", left, right)
		} else {
			return self.builder.CreateURem("", left, right)
		}
	case *local.LtExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lt")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFloatCmp("", llvm.FloatOLT, left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateIntCmp("", llvm.IntSLT, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntULT, left, right)
		}
	case *local.GtExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Gt")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFloatCmp("", llvm.FloatOGT, left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateIntCmp("", llvm.IntSGT, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntUGT, left, right)
		}
	case *local.LeExpr:
		eqTrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		ltTrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lt")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && eqTrait.HasBeImpled(lt, true) && ltTrait.HasBeImpled(lt, true) {
			eqMethod := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(eqTrait.FirstMethod()).GetName()).Value))
			leMethod := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(ltTrait.FirstMethod()).GetName()).Value))
			return self.codegenBinary(local.NewLogicAndExpr(
				local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), eqMethod, nil), ir.GetRight()),
				local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), leMethod, nil), ir.GetRight()),
			), load)
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFloatCmp("", llvm.FloatOLE, left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateIntCmp("", llvm.IntSLE, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntULE, left, right)
		}
	case *local.GeExpr:
		eqTrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		gtTrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Gt")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && eqTrait.HasBeImpled(lt, true) && gtTrait.HasBeImpled(lt, true) {
			eqMethod := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(eqTrait.FirstMethod()).GetName()).Value))
			gtMethod := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(gtTrait.FirstMethod()).GetName()).Value))
			return self.codegenBinary(local.NewLogicAndExpr(
				local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), eqMethod, nil), ir.GetRight()),
				local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), gtMethod, nil), ir.GetRight()),
			), load)
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		if types.Is[types.FloatType](lt) {
			return self.builder.CreateFloatCmp("", llvm.FloatOGE, left, right)
		} else if types.Is[types.SintType](lt) {
			return self.builder.CreateIntCmp("", llvm.IntSGE, left, right)
		} else {
			return self.builder.CreateIntCmp("", llvm.IntUGE, left, right)
		}
	case *local.EqExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight()))
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateZExt("", self.buildEqual(ir.GetLeft().Type(), left, right, false), self.builder.BooleanType())
	case *local.NeExpr:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](lt, true); ok && trait.HasBeImpled(lt, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenUnary(local.NewNotExpr(local.NewCallExpr(local.NewMethodExpr(ir.GetLeft(), method, nil), ir.GetRight())), load)
		}
		left, right := self.codegenValue(ir.GetLeft(), true), self.codegenValue(ir.GetRight(), true)
		return self.builder.CreateZExt("", self.buildEqual(ir.GetLeft().Type(), left, right, true), self.builder.BooleanType())
	case *local.IndexExpr:
		return self.codegenIndex(ir, load)
	case *local.ExtractExpr:
		return self.codegenExtract(ir, load)
	case *local.FieldExpr:
		return self.codegenField(ir, load)
	case *local.MethodExpr:
		return self.codegenMethod(ir)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenUnary(ir local.UnaryExpr, load bool) llvm.Value {
	switch ir := ir.(type) {
	case *local.NegExpr:
		tIr := ir.Type()
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Neg")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](tIr, true); ok && trait.HasBeImpled(tIr, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetValue(), method, nil)))
		}
		return self.codegenBinary(local.NewSubExpr(
			local.NewDefaultExpr(ir.GetValue().Type()),
			ir.GetValue(),
		), load)
	case *local.NotExpr:
		tIr := ir.Type()
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Not")).(*global.Trait)
		if ctd, ok := types.As[global.CustomTypeDef](tIr, true); ok && trait.HasBeImpled(tIr, true) {
			method := stlval.IgnoreWith(ctd.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName()).Value))
			return self.codegenCall(local.NewCallExpr(local.NewMethodExpr(ir.GetValue(), method, nil)))
		}
		return self.builder.CreateNot("", self.codegenValue(ir.GetValue(), true))
	case *local.GetRefExpr:
		return self.codegenValue(ir.GetValue(), false)
	case *local.DeRefExpr:
		ptr := self.codegenValue(ir.GetValue(), true)
		if !load {
			return ptr
		}
		return self.builder.CreateLoad("", self.codegenType(ir.Type()), ptr)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIdent(ir values.Ident, load bool) llvm.Value {
	if !self.lambdaCaptureMap.Empty() {
		if mapVal := self.lambdaCaptureMap.Peek().Get(ir); mapVal != nil {
			if !load {
				return mapVal
			}
			return self.builder.CreateLoad("", self.codegenType(ir.Type()), mapVal)
		}
	}

	switch identIr := ir.(type) {
	case *global.FuncDef:
		if !self.pkg.Equal(identIr.Package()) {
			name := self.getIdentName(identIr)
			f, ok := self.builder.GetFunction(name)
			if ok {
				return f
			}
			f = self.declFunc(name, identIr.CallableType().(types.FuncType), identIr.Attrs()...)
			f.SetLinkage(llvm.LinkOnceODRAutoHideLinkage)
			return f
		}
		return self.values.Get(identIr)
	case global.MethodDef:
		if !self.pkg.Equal(identIr.Package()) {
			name := self.getIdentName(identIr)
			f, ok := self.builder.GetFunction(name)
			if ok {
				return f
			}
			f = self.declFunc(name, identIr.CallableType().(types.FuncType), identIr.Attrs()...)
			f.SetLinkage(llvm.LinkOnceODRAutoHideLinkage)
			return f
		}
		return self.values.Get(identIr)
	case *global.VarDef:
		p := stlval.TernaryAction(!self.pkg.Equal(identIr.Package()), func() llvm.Value {
			name := self.getIdentName(identIr)
			v, ok := self.builder.GetGlobal(name)
			if ok {
				return v
			}
			v = self.builder.NewGlobal(name, self.codegenType(identIr.Type()), nil)
			v.SetLinkage(llvm.LinkOnceODRAutoHideLinkage)
			return v
		}, func() llvm.Value {
			return self.values.Get(identIr)
		})
		if !load {
			return p
		}
		return self.builder.CreateLoad("", self.codegenType(identIr.Type()), p)
	case *local.SingleVarDef, values.VarDecl, *local.Param:
		p := self.values.Get(identIr)
		if !load {
			return p
		}
		return self.builder.CreateLoad("", self.codegenType(identIr.Type()), p)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenCall(ir *local.CallExpr) llvm.Value {
	if stlval.Is[*local.MethodExpr](ir.GetFunc()) {
		return self.codegenMethodCall(ir)
	} else if stlval.Is[*local.TraitMethodExpr](ir.GetFunc()) {
		return self.codegenTraitMethodCall(ir)
	} else if types.Is[types.LambdaType](ir.GetFunc().Type()) {
		return self.codegenLambdaCall(ir)
	} else {
		return self.codegenFuncCall(ir)
	}
}

func (self *CodeGenerator) codegenFuncCall(ir *local.CallExpr) llvm.Value {
	fIr := ir.GetFunc()
	f := self.codegenValue(fIr, true)
	args := stlslices.Map(ir.GetArgs(), func(_ int, argIr hir.Value) llvm.Value {
		return self.codegenValue(argIr, true)
	})
	ft := self.codegenFuncType(fIr.Type().(types.FuncType))
	if fdefIr, ok := fIr.(*global.FuncDef); ok {
		vararg := stlslices.Exist(fdefIr.Attrs(), func(_ int, attr global.FuncAttr) bool {
			return stlval.Is[*global.FuncAttrVararg](attr)
		})
		if vararg {
			ft = self.builder.FunctionType(true, ft.ReturnType(), ft.Params()...)
		}
	}
	return self.builder.CreateCall("", ft, f, args...)
}

func (self *CodeGenerator) codegenMethodCall(ir *local.CallExpr) llvm.Value {
	methodIr := ir.GetFunc().(*local.MethodExpr)

	selfVal := self.codegenValue(methodIr.GetLeft(), true)
	args := stlslices.Map(ir.GetArgs(), func(_ int, argIr hir.Value) llvm.Value {
		return self.codegenValue(argIr, true)
	})

	var method llvm.Value
	var ft llvm.FunctionType
	if methodIr.GenericParamMap().Length() > 0 {
		method = self.instGenericFunc(methodIr.Method().(global.MethodDef), methodIr)
		ft = method.(llvm.Function).FunctionType()
	} else {
		method = self.codegenValue(methodIr.Method(), true)
		ft = self.codegenFuncType(methodIr.Method().CallableType().(types.FuncType))
	}

	return self.builder.CreateCall("", ft, method, append([]llvm.Value{selfVal}, args...)...)
}

func (self *CodeGenerator) codegenTraitMethodCall(ir *local.CallExpr) llvm.Value {
	methodIr := ir.GetFunc().(*local.TraitMethodExpr)

	selfType := methodIr.Self().Type()
	trait := stlval.IgnoreWith(stlval.IgnoreWith(types.As[types.GenericParamType](selfType)).Restraint()).(*global.Trait)
	selfType = types.ReplaceVirtualType(self.virtualTypes, selfType)
	if rt, ok := types.As[types.RefType](selfType, true); ok {
		selfType = rt.Pointer()
	}

	if trait.HasBeImpled(selfType, true) {
		return self.codegenMethodCall(local.NewCallExpr(self.covertTraitMethodToMethod(methodIr), ir.GetArgs()...))
	}

	return self.codegenValue(stlval.IgnoreWith(trait.GetCovertValue(methodIr.Self(), ir.GetArgs()...)), true)
}

func (self *CodeGenerator) codegenLambdaCall(ir *local.CallExpr) llvm.Value {
	lts := self.codegenLambdaType()
	ltIr := stlval.IgnoreWith(types.As[types.LambdaType](ir.GetFunc().Type()))
	ft := self.codegenFuncType(ltIr.ToFunc())
	lt := self.builder.FunctionType(ft.IsVarArg(), ft.ReturnType(), append([]llvm.Type{self.builder.OpaquePointerType()}, ft.Params()...)...)
	f := self.codegenValue(ir.GetFunc(), true)
	args := stlslices.Map(ir.GetArgs(), func(_ int, argIr hir.Value) llvm.Value {
		return self.codegenValue(argIr, true)
	})

	fPtr := self.builder.CreateStructIndex(lts, f, 0, false)
	ctxPtr := self.builder.CreateStructIndex(lts, f, 1, false)
	cf := self.builder.CurrentFunction()
	f1block, f2block, endblock := cf.NewBlock(""), cf.NewBlock(""), cf.NewBlock("")
	self.builder.CreateCondBr(self.builder.CreateIntCmp("", llvm.IntEQ, ctxPtr, self.builder.ConstZero(ctxPtr.Type())), f1block, f2block)

	self.builder.MoveToAfter(f1block)
	f1ret := self.builder.CreateCall("", ft, fPtr, args...)
	self.builder.CreateBr(endblock)

	self.builder.MoveToAfter(f2block)
	f2ret := self.builder.CreateCall("", lt, fPtr, append([]llvm.Value{ctxPtr}, args...)...)
	self.builder.CreateBr(endblock)

	self.builder.MoveToAfter(endblock)
	if f1ret.Type().Equal(self.builder.VoidType()) {
		return f1ret
	} else {
		return self.builder.CreatePHI(
			"",
			f1ret.Type(),
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: f1ret, Block: f1block},
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: f2ret, Block: f2block},
		)
	}
}

func (self *CodeGenerator) codegenCovert(ir local.CovertExpr, load bool) llvm.Value {
	switch ir.(type) {
	case *local.WrapTypeExpr, *local.Enum2NumberExpr, *local.Number2EnumExpr:
		return self.codegenValue(ir.GetFrom(), load)
	case *local.Int2IntExpr:
		ftIr := ir.GetFrom().Type()
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type())
		ftSize, toSize := from.Type().(llvm.IntegerType).Bits(), tt.(llvm.IntegerType).Bits()
		if ftSize < toSize {
			if types.Is[types.SignedType](ftIr) {
				return self.builder.CreateSExt("", from, tt.(llvm.IntegerType))
			} else {
				return self.builder.CreateZExt("", from, tt.(llvm.IntegerType))
			}
		} else if ftSize > toSize {
			return self.builder.CreateTrunc("", from, tt.(llvm.IntegerType))
		} else {
			return from
		}
	case *local.Int2FloatExpr:
		ftIr := ir.GetFrom().Type()
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type())
		if types.Is[types.SignedType](ftIr) {
			return self.builder.CreateSIToFP("", from, tt.(llvm.FloatType))
		} else {
			return self.builder.CreateUIToFP("", from, tt.(llvm.FloatType))
		}
	case *local.Float2IntExpr:
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type())
		if types.Is[types.SignedType](ir.Type()) {
			return self.builder.CreateFPToSI("", from, tt.(llvm.IntegerType))
		} else {
			return self.builder.CreateFPToUI("", from, tt.(llvm.IntegerType))
		}
	case *local.Float2FloatExpr:
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type())
		ftSize, toSize := from.Type().(llvm.FloatType).Kind(), tt.(llvm.FloatType).Kind()
		if ftSize < toSize {
			return self.builder.CreateFPExt("", from, tt.(llvm.FloatType))
		} else if ftSize > toSize {
			return self.builder.CreateFPTrunc("", from, tt.(llvm.FloatType))
		} else {
			return from
		}
	case *local.Ref2UsizeExpr:
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type()).(llvm.IntegerType)
		return self.builder.CreatePtrToInt("", from, tt)
	case *local.Usize2RefExpr:
		from := self.codegenValue(ir.GetFrom(), true)
		tt := self.codegenType(ir.Type()).(llvm.PointerType)
		return self.builder.CreateIntToPtr("", from, tt)
	case *local.NoReturn2AnyExpr:
		from := self.codegenValue(ir.GetFrom(), false)
		self.builder.CreateUnreachable()
		if types.Is[types.NoThingType](ir.Type()) {
			return from
		} else {
			return self.builder.ConstZero(self.codegenType(ir.Type()))
		}
	case *local.Func2LambdaExpr:
		t := self.codegenType(ir.Type()).(llvm.StructType)
		f := self.codegenValue(ir.GetFrom(), true)
		return self.builder.CreateStruct(t, f, self.builder.ConstZero(t.Elems()[1]), self.builder.ConstZero(t.Elems()[2]))
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenArray(ir *local.ArrayExpr) llvm.Value {
	elems := stlslices.Map(ir.Elems(), func(_ int, elem hir.Value) llvm.Value {
		return self.codegenValue(elem, true)
	})
	return self.builder.CreatePackArray(self.codegenType(ir.Type()).(llvm.ArrayType), elems...)
}

func (self *CodeGenerator) codegenIndex(ir *local.IndexExpr, load bool) llvm.Value {
	at := self.codegenType(ir.GetLeft().Type()).(llvm.ArrayType)
	index := self.codegenValue(ir.GetRight(), true)
	self.buildCheckIndex(index, uint64(at.Capacity()))
	from := self.codegenValue(ir.GetLeft(), false)
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.builder.CreateArrayIndex(at, from, index, expectPtr...)
}

func (self *CodeGenerator) codegenTuple(ir *local.TupleExpr) llvm.Value {
	elems := stlslices.Map(ir.Elems(), func(_ int, elem hir.Value) llvm.Value {
		return self.codegenValue(elem, true)
	})
	return self.builder.CreateStruct(self.codegenType(ir.Type()).(llvm.StructType), elems...)
}

func (self *CodeGenerator) codegenExtract(ir *local.ExtractExpr, load bool) llvm.Value {
	from := self.codegenValue(ir.GetLeft(), false)
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.builder.CreateStructIndex(self.codegenType(ir.GetLeft().Type()).(llvm.StructType), from, ir.Index(), expectPtr...)
}

func (self *CodeGenerator) codegenStruct(ir *local.StructExpr) llvm.Value {
	fields := stlslices.Map(ir.Fields(), func(_ int, field hir.Value) llvm.Value {
		return self.codegenValue(field, true)
	})
	return self.builder.CreateStruct(self.codegenType(ir.Type()).(llvm.StructType), fields...)
}

func (self *CodeGenerator) codegenField(ir *local.FieldExpr, load bool) llvm.Value {
	from := self.codegenValue(ir.GetLeft(), false)
	stIr := stlval.IgnoreWith(types.As[types.StructType](ir.GetLeft().Type()))
	var index uint
	for iter := stIr.Fields().Iterator(); iter.Next(); index++ {
		if iter.Value().E2().Name() == ir.Field() {
			break
		}
	}
	var expectPtr []bool
	if load {
		expectPtr = []bool{false}
	}
	return self.builder.CreateStructIndex(self.codegenType(ir.GetLeft().Type()).(llvm.StructType), from, index, expectPtr...)
}

func (self *CodeGenerator) codegenTypeJudgment(ir *local.TypeJudgmentExpr) llvm.Value {
	return self.builder.CreateZExt("", self.builder.ConstBoolean(ir.Value().Type().Equal(ir.Type())), self.builder.BooleanType())
}

func (self *CodeGenerator) codegenLambda(ir *local.LambdaExpr) llvm.Value {
	lts := self.codegenLambdaType()
	ft := self.codegenFuncType(ir.CallableType().ToFunc())
	lt := self.builder.FunctionType(ft.IsVarArg(), ft.ReturnType(), append([]llvm.Type{self.builder.OpaquePointerType()}, ft.Params()...)...)
	isSimpleFunc := len(ir.Context()) == 0
	vft := stlval.Ternary(isSimpleFunc, ft, lt)
	f := self.builder.NewFunction("", vft)
	defer self.removeUnreachableInst(f)
	if types.Is[types.NoReturnType](ir.CallableType().Ret()) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}

	if isSimpleFunc {
		preBlock := self.builder.CurrentBlock()
		self.builder.MoveToAfter(f.NewBlock(""))
		for i, pir := range ir.Params() {
			p := self.builder.CreateAlloca("", self.codegenType(pir.Type()))
			self.builder.CreateStore(f.GetParam(uint(i)), p)
			self.values.Set(pir, p)
		}

		block, _ := self.codegenBlock(stlval.IgnoreWith(ir.Body()), nil)
		self.builder.CreateBr(block)
		self.builder.MoveToAfter(preBlock)

		return self.builder.CreateStruct(lts, f, self.builder.ConstZero(lts.GetElem(1)))
	} else {
		ctxType := self.builder.StructType(false, stlslices.Map(ir.Context(), func(_ int, e values.Ident) llvm.Type {
			return self.builder.OpaquePointerType()
		})...)
		externalCtxPtr := self.buildMalloc(ctxType)
		for i, identIr := range ir.Context() {
			self.builder.CreateStore(
				self.codegenIdent(identIr, false),
				self.builder.CreateStructIndex(ctxType, externalCtxPtr, uint(i), true),
			)
		}

		preBlock := self.builder.CurrentBlock()
		self.builder.MoveToAfter(f.NewBlock(""))
		for i, pir := range ir.Params() {
			p := self.builder.CreateAlloca("", self.codegenType(pir.Type()))
			self.builder.CreateStore(f.GetParam(uint(i+1)), p)
			self.values.Set(pir, p)
		}

		captureMap := hashmap.StdWithCap[values.Ident, llvm.Value](uint(len(ir.Context())))
		for i, identIr := range ir.Context() {
			captureMap.Set(identIr, self.builder.CreateStructIndex(ctxType, f.GetParam(0), uint(i), false))
		}

		self.lambdaCaptureMap.Push(captureMap)
		defer func() {
			self.lambdaCaptureMap.Pop()
		}()

		block, _ := self.codegenBlock(stlval.IgnoreWith(ir.Body()), nil)
		self.builder.CreateBr(block)
		self.builder.MoveToAfter(preBlock)

		return self.builder.CreateStruct(lts, f, externalCtxPtr)
	}
}

func (self *CodeGenerator) codegenMethod(ir *local.MethodExpr) llvm.Value {
	var method llvm.Value
	if ir.GenericParamMap().Length() > 0 {
		method = self.instGenericFunc(ir.Method().(*global.OriginMethodDef), ir)
	} else {
		method = self.codegenIdent(ir.Method().(values.Ident), true)
	}

	lts := self.codegenLambdaType()
	ft := self.codegenFuncType(ir.CallableType().ToFunc())
	lt := self.builder.FunctionType(ft.IsVarArg(), ft.ReturnType(), append([]llvm.Type{self.builder.OpaquePointerType()}, ft.Params()...)...)
	f := self.builder.NewFunction("", lt)
	if types.Is[types.NoReturnType](ir.Method().CallableType().Ret(), true) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}

	ctxType := self.builder.StructType(false, self.codegenType(ir.GetLeft().Type()))
	externalCtxPtr := self.buildMalloc(ctxType)
	self.builder.CreateStore(
		self.codegenValue(ir.GetLeft(), true),
		self.builder.CreateStructIndex(ctxType, externalCtxPtr, 0, true),
	)

	preBlock := self.builder.CurrentBlock()
	self.builder.MoveToAfter(f.NewBlock(""))
	selfVal := self.builder.CreateStructIndex(ctxType, f.GetParam(0), 0, false)
	args := []llvm.Value{selfVal}
	args = append(args, stlslices.Map(f.Params()[1:], func(i int, e llvm.Param) llvm.Value {
		return e
	})...)
	ret := self.builder.CreateCall("", self.codegenFuncType(ir.Method().CallableType().ToFunc()), method, args...)
	if ret.Type().Equal(self.builder.VoidType()) {
		self.builder.CreateRet(nil)
	} else {
		self.builder.CreateRet(ret)
	}

	self.builder.MoveToAfter(preBlock)
	return self.builder.CreateStruct(lts, f, externalCtxPtr)
}

func (self *CodeGenerator) codegenEnum(ir *local.EnumExpr) llvm.Value {
	etIr := stlval.IgnoreWith(types.As[types.EnumType](ir.Type()))
	index := slices.Index(etIr.EnumFields().Keys(), ir.Field())
	if etIr.Simple() {
		return self.builder.ConstInteger(self.codegenType(etIr).(llvm.IntegerType), int64(index))
	}

	ut := self.codegenType(ir.Type()).(llvm.StructType)
	ptr := self.builder.CreateAlloca("", ut)
	if elemIr, ok := ir.Elem(); ok {
		value := self.codegenValue(elemIr, true)
		self.builder.CreateStore(value, self.builder.CreateStructIndex(ut, ptr, 0, true))
	}
	self.builder.CreateStore(
		self.builder.ConstInteger(ut.GetElem(1).(llvm.IntegerType), int64(index)),
		self.builder.CreateStructIndex(ut, ptr, 1, true),
	)
	return self.builder.CreateLoad("", ut, ptr)
}

func (self *CodeGenerator) codegenGenericFuncInst(ir *local.GenericFuncInstExpr) llvm.Value {
	return self.instGenericFunc(ir.GetFunc().(*global.FuncDef), ir)
}

func (self *CodeGenerator) codegenStaticMethod(ir *local.StaticMethodExpr) llvm.Value {
	if ir.GenericParamMap().Length() == 0 {
		return self.codegenIdent(ir.Method().(values.Ident), true)
	}

	return self.instGenericFunc(ir.Method().(*global.OriginMethodDef), ir)
}

func (self *CodeGenerator) codegenTraitStaticMethod(ir *local.TraitStaticMethodExpr) llvm.Value {
	selfType := types.ReplaceVirtualType(self.virtualTypes, ir.Self()).(global.CustomTypeDef)
	method := stlval.IgnoreWith(selfType.GetMethod(ir.Method()))
	return self.codegenStaticMethod(local.NewStaticMethodExpr(selfType, method, nil))
}

func (self *CodeGenerator) covertTraitMethodToMethod(ir *local.TraitMethodExpr) hir.Value {
	selfType := ir.Self().Type()
	trait := stlval.IgnoreWith(stlval.IgnoreWith(types.As[types.GenericParamType](selfType)).Restraint()).(*global.Trait)
	selfType = types.ReplaceVirtualType(self.virtualTypes, selfType)
	if rt, ok := types.As[types.RefType](selfType, true); ok {
		selfType = rt.Pointer()
	}

	if trait.HasBeImpled(selfType, true) {
		selfGpt := stlval.IgnoreWith(types.As[types.GenericParamType](selfType, true))
		selfCtd := self.virtualTypes.Get(selfGpt).(global.CustomTypeDef)
		method := stlval.IgnoreWith(selfCtd.GetMethod(ir.Method()))
		return local.NewMethodExpr(ir.Self(), method, nil)
	}

	panic("unreachable")
}
