package analyse

import (
	"math/big"

	"github.com/samber/lo"

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
	params := lo.Map(node.Params, func(item ast.Type, index int) hir.Type {
		return self.analyseType(item)
	})
	return &hir.FuncType{
		Ret:    self.analyseOptionType(node.Ret),
		Params: params,
	}
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *hir.ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return &hir.ArrayType{
		Size: size.Uint64(),
		Elem: elem,
	}
}

func (self *Analyser) analyseTupleType(node *ast.TupleType) *hir.TupleType {
	elems := lo.Map(node.Elems, func(item ast.Type, index int) hir.Type {
		return self.analyseType(item)
	})
	return &hir.TupleType{Elems: elems}
}

func (self *Analyser) analyseUnionType(node *ast.UnionType) *hir.UnionType {
	return &hir.UnionType{Elems: lo.Map(node.Elems.ToSlice(), func(item ast.Type, _ int) hir.Type {
		return self.analyseType(item)
	})}
}

func (self *Analyser) analyseRefType(node *ast.RefType) *hir.RefType {
	return &hir.RefType{Elem: self.analyseType(node.Elem)}
}

func (self *Analyser) analyseSelfType(node *ast.SelfType) hir.Type{
	if self.selfType == nil && !self.inTrait{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}else if self.inTrait{
		return &hir.SelfType{Self: util.None[hir.TypeDef]()}
	}
	return &hir.SelfType{Self: util.Some(self.selfType)}
}
