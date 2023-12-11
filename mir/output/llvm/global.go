package llvm

import (
	"github.com/kkkunny/go-llvm"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

func (self *LLVMOutputer) codegenDeclType(ir mir.Global){
	switch global := ir.(type) {
	case *mir.NamedStruct:
		self.types.Set(global, self.ctx.NamedStructType(global.RealName(), false))
	case *mir.GlobalVariable, *mir.Constant, *mir.Function:
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenDefType(ir mir.Global){
	switch global := ir.(type) {
	case *mir.NamedStruct:
		st := self.types.Get(global).(llvm.StructType)
		elems := lo.Map(global.Elems(), func(item mir.Type, _ int) llvm.Type {
			return self.codegenType(item)
		})
		st.SetElems(false, elems...)
		self.types.Set(global, st)
	case *mir.GlobalVariable, *mir.Constant, *mir.Function:
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenDeclValue(ir mir.Global){
	switch global := ir.(type) {
	case *mir.NamedStruct:
	case *mir.GlobalVariable:
		g := self.module.NewGlobal(global.RealName(), self.codegenType(global.ValueType()))
		if global.RealName() == ""{
			g.SetLinkage(llvm.InternalLinkage)
		}else{
			g.SetLinkage(llvm.ExternalLinkage)
		}
		g.SetAlign(uint32(global.ValueType().Align()))
		self.values.Set(global, g)
	case *mir.Constant:
		g := self.module.NewGlobal(global.RealName(), self.codegenType(global.ValueType()))
		if global.RealName() == ""{
			g.SetLinkage(llvm.InternalLinkage)
		}else{
			g.SetLinkage(llvm.ExternalLinkage)
		}
		g.SetGlobalConstant(true)
		g.SetAlign(uint32(global.ValueType().Align()))
		self.values.Set(global, g)
	case *mir.Function:
		f := self.module.NewFunction(global.RealName(), self.codegenFuncType(global.Type().(mir.FuncType)))
		if global.RealName() == ""{
			f.SetLinkage(llvm.InternalLinkage)
		}else{
			f.SetLinkage(llvm.ExternalLinkage)
		}
		self.values.Set(global, f)
	default:
		panic("unreachable")
	}
}

func (self *LLVMOutputer) codegenDefValue(ir mir.Global){
	switch global := ir.(type) {
	case *mir.NamedStruct:
	case *mir.GlobalVariable:
		g := self.values.Get(global).(llvm.GlobalValue)
		if v, ok := global.Value(); ok{
			g.SetInitializer(self.codegenConst(v))
		}
	case *mir.Constant:
		g := self.values.Get(global).(llvm.GlobalValue)
		g.SetInitializer(self.codegenConst(global.Value()))
	case *mir.Function:
		f := self.values.Get(global).(llvm.Function)
		switch {
		case self.target.IsWindows() && self.codegenType(global.Type().(mir.FuncType).Ret()).String() != f.FunctionType().ReturnType().String():
			for i, paramIr := range global.Params(){
				self.values.Set(paramIr, f.GetParam(uint(i)+1))
			}
		default:
			for i, paramIr := range global.Params(){
				self.values.Set(paramIr, f.GetParam(uint(i)))
			}
		}
		if blockIrs := global.Blocks(); !blockIrs.Empty(){
			for iter:=blockIrs.Iterator(); iter.Next(); {
				self.blocks.Set(iter.Value(), f.NewBlock(""))
			}
		}
		if blockIrs := global.Blocks(); !blockIrs.Empty(){
			for iter:=blockIrs.Iterator(); iter.Next(); {
				self.codegenBlock(iter.Value())
			}
		}
	default:
		panic("unreachable")
	}
}
