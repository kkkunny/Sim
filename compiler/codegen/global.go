package codegen

import (
	"fmt"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genTypeDecl(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.TypeDef:
		c.genCustomTypeDecl(global)
	}
}

func (c *CodeGenerator) genCustomTypeDecl(global *stmts.TypeDef) {
	switch global.Type.GetUnderlying().(type) {
	case types.TupleType, types.ArrayType, types.UnionType, types.FuncType, types.RefType, types.StructType:
	default:
		return
	}

	name := stableName(c.pkg, global.Type.GetName())
	st := cir.NewStructType(name, optional.None[[]*cir.Member]())
	def := cir.BuildStmt(c.builder, cir.NewTypedef(st, name))
	st.Name = def.Name
	c.ctx.typeCache[global.Type.GetName()] = cir.NewAliasType(def)
}

func (c *CodeGenerator) genTypeDef(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.TypeDef:
		c.genCustomTypeDef(global.Type)
	}
}

func (c *CodeGenerator) genCustomTypeDef(ct types.CustomType) cir.Type {
	t, ok := c.ctx.typeCache[ct.GetName()]
	switch underlyingHir := ct.GetUnderlying().(type) {
	case types.TupleType:
		st := c.genFlatTupleType(underlyingHir)
		st.Name = t.Def.Name
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(st))
	case types.ArrayType:
		at := cir.NewStructType(t.Def.Name, optional.Some([]*cir.Member{
			cir.NewMember(cir.NewArrayType(c.genType(underlyingHir.GetElem()), underlyingHir.GetSize()), "array"),
		}))
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(at))
	case types.UnionType:
		ut := c.genFlatUnionType(underlyingHir)
		ut.Name = t.Def.Name
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(ut))
	case types.FuncType:
		r := c.genType(underlyingHir.GetReturn())
		ps := stlslices.Map(underlyingHir.GetParams(), func(i int, e types.Type) cir.Type {
			return c.genType(e)
		})
		ft := cir.NewStructType(t.Def.Name, optional.Some([]*cir.Member{
			cir.NewMember(cir.NewUnionType(
				"",
				cir.NewMember(cir.NewPointerType(cir.NewFuncType(r, ps...)), "f"),
				cir.NewMember(cir.NewPointerType(cir.NewFuncType(r, append([]cir.Type{cir.VoidPtr}, ps...)...)), "c"),
			), "func"),
			cir.NewMember(cir.VoidPtr, "ctx"),
		}))
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(ft))
	case types.RefType:
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(cir.NewStructType(t.Def.Name, optional.Some([]*cir.Member{
			cir.NewMember(cir.NewPointerType(c.genType(underlyingHir.PtrTo())), "ptr"),
		}))))
	case types.StructType:
		st := c.genStructType(underlyingHir)
		st.Name = t.Def.Name
		cir.BuildStmt(c.builder, cir.NewStructTypeDef(st))
	default:
		if ok {
			return t
		}
		underlying := c.genType(ct.GetUnderlying())
		name := stableName(c.pkg, ct.GetName())
		def := cir.BuildStmt(c.builder, cir.NewTypedef(underlying, name))
		c.ctx.typeCache[ct.GetName()] = cir.NewAliasType(def)
	}
	return c.ctx.typeCache[ct.GetName()]
}

func (c *CodeGenerator) genGlobalValue(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.Let:
		c.genGlobalLet(global)
	}
}

func (c *CodeGenerator) genGlobalLet(l *stmts.Let) {
	if !l.Mut && stlval.Is[types.FuncType](l.GetType()) {
		if l.Value.IsSome() && stlval.Is[*stmts.Func](l.Value.MustValue()) {
			expr := l.Value.MustValue().(*stmts.Func)
			decl := c.genNativeFuncDecl(expr)
			if l.Name == "main" {
				decl.Name = "sim_main"
				decl.Static = true
			} else if externalName, ok := l.ExternalName.Value(); ok {
				decl.Name = externalName
			} else {
				decl.Name = stableName(c.pkg, l.Name)
				decl.Static = !l.Pub
			}
			c.ctx.idents[l] = &Ident{Name: decl.GetName()}
			if b, ok := expr.Body.Value(); ok {
				prevFunc := c.currentFunc
				c.currentFunc = expr
				decl.Body = optional.Some(c.genFuncBlock(b, nil))
				c.currentFunc = prevFunc
			}
			return
		} else if l.Value.IsNone() {
			ftHir := l.GetType().(types.FuncType)
			rt := c.genType(ftHir.GetReturn())
			params := stlslices.Map(ftHir.GetParams(), func(i int, p types.Type) *cir.Param {
				return cir.NewParam(fmt.Sprintf("_p%d", i+1), c.genType(p))
			})
			decl := cir.BuildStmt(c.builder, cir.NewFunc(l.ExternalName.MustValue(), rt, params...))
			c.ctx.idents[l] = &Ident{Name: decl.GetName(), ExternalFunc: true}
			return
		}
	}

	t := c.genType(l.GetType())
	pub := l.Pub
	var name string
	if externalName, ok := l.ExternalName.Value(); ok {
		pub = true
		name = externalName
	} else {
		name = stableName(c.pkg, l.Name)
	}
	decl := cir.BuildStmt(c.builder, cir.NewVariable(t, name))
	decl.Static = !pub
	c.ctx.idents[l] = &Ident{Name: decl.GetName()}
	if v, ok := l.Value.Value(); ok {
		decl.Value = optional.Some(c.genExpr(v))
	}
}
