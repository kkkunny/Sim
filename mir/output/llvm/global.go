package llvm

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
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
		for _, attr := range global.Attributes(){
			switch attr {
			case mir.FunctionAttributeInit:
				self.module.AddConstructor(65535, f)
			case mir.FunctionAttributeFini:
				self.module.AddDestructor(65535, f)
			case mir.FunctionAttributeNoReturn:
				f.AddAttribute(llvm.FuncAttributeNoReturn)
			default:
				panic("unreachable")
			}
		}
		f.AddAttribute(llvm.FuncAttributeInlineHint)
		f.AddAttribute(llvm.FuncAttributeAllocKind, 1|16|32)
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
		if blockIrs := global.Blocks(); blockIrs.Len() == 0{
			return
		}

		self.builder.MoveToAfter(f.NewBlock("entry"))
		params := f.Params()
		if self.target.IsWindows() && self.codegenType(global.Type().(mir.FuncType).Ret()).String() != f.FunctionType().ReturnType().String(){
			params = params[1:]
		}
		for i, param := range params{
			srcParamType := self.codegenType(global.Params()[i].ValueType())
			if self.target.IsWindows() && (stlbasic.Is[llvm.StructType](srcParamType) || stlbasic.Is[llvm.ArrayType](srcParamType)){
				self.values.Set(global.Params()[i], param)
			}else{
				ptr := self.builder.CreateAlloca("", param.Type())
				self.builder.CreateStore(param, ptr)
				self.values.Set(global.Params()[i], ptr)
			}
		}

		for cursor:=global.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
			self.blocks.Set(cursor.Value, f.NewBlock(""))
		}
		self.builder.CreateBr(self.blocks.Get(global.Blocks().Front().Value))
		for cursor:=global.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
			self.codegenBlock(cursor.Value)
		}
	default:
		panic("unreachable")
	}
}
