package analyse

import (
	"math/big"

	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseType(node ast.Type) hir.Type {
	switch typeNode := node.(type) {
	case *ast.IdentType:
		return self.analyseIdentType(typeNode)
	case *ast.FuncType:
		return self.analyseFuncType(typeNode)
	case *ast.ArrayType:
		return self.analyseArrayType(typeNode)
	case *ast.TupleType:
		return self.analyseTupleType(typeNode)
	case *ast.UnionType:
		return self.analyseUnionType(typeNode)
	case *ast.RefType:
		return self.analyseRefType(typeNode)
	case *ast.SelfType:
		return self.analyseSelfType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) hir.Type {
	t, ok := node.Value()
	if !ok {
		return hir.Empty
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) hir.Type {
	typ := self.analyseIdent((*ast.Ident)(node), false)
	if typ.IsNone(){
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	t, _ := typ.MustValue().Right()
	return t
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *hir.FuncType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e)
	})
	return hir.NewFuncType(self.analyseOptionType(node.Ret), params...)
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *hir.ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return hir.NewArrayType(size.Uint64(), elem)
}

func (self *Analyser) analyseTupleType(node *ast.TupleType) *hir.TupleType {
	elems := stlslices.Map(node.Elems, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e)
	})
	return hir.NewTupleType(elems...)
}

func (self *Analyser) analyseUnionType(node *ast.UnionType) *hir.UnionType {
	elems := stlslices.Map(node.Elems.ToSlice(), func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e)
	})
	return hir.NewUnionType(elems...)
}

func (self *Analyser) analyseRefType(node *ast.RefType) *hir.RefType {
	return hir.NewRefType(node.Mut, self.analyseType(node.Elem))
}

func (self *Analyser) analyseSelfType(node *ast.SelfType) hir.Type{
	if self.selfType == nil{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}
	return hir.NewSelfType(self.selfType)
}
