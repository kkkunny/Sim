package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/hir"

	errors "github.com/kkkunny/Sim/compiler/error"
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
	case *ast.RefType:
		return self.analyseRefType(typeNode)
	case *ast.SelfType:
		return self.analyseSelfType(typeNode)
	case *ast.StructType:
		return self.analyseStructType(typeNode)
	case *ast.LambdaType:
		return self.analyseLambdaType(typeNode)
	case *ast.EnumType:
		return self.analyseEnumType(typeNode)
	default:
		panic("unreachable")
	}
}

var voidTypeAnalyser = func(node optional.Optional[ast.Type]) optional.Optional[hir.Type] {
	typeNode, ok := node.Value()
	if !ok {
		return optional.None[hir.Type]()
	}
	if ident, ok := typeNode.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == hir.NoThing.String() {
		return optional.Some[hir.Type](hir.NoThing)
	}
	return optional.None[hir.Type]()
}

var noReturnTypeAnalyser = func(node optional.Optional[ast.Type]) optional.Optional[hir.Type] {
	typeNode, ok := node.Value()
	if !ok {
		return optional.None[hir.Type]()
	}
	if ident, ok := typeNode.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == hir.NoReturn.String() {
		return optional.Some[hir.Type](hir.NoReturn)
	}
	return optional.None[hir.Type]()
}

func (self *Analyser) analyseOptionType(node optional.Optional[ast.Type]) hir.Type {
	t, ok := node.Value()
	if !ok {
		return hir.NoThing
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseOptionTypeWith(node optional.Optional[ast.Type], analysers ...func(node optional.Optional[ast.Type]) optional.Optional[hir.Type]) hir.Type {
	for _, analyser := range analysers {
		if t, ok := analyser(node).Value(); ok {
			return t
		}
	}
	return self.analyseOptionType(node)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) hir.Type {
	typ := self.analyseIdent((*ast.Ident)(node), false)
	if typ.IsNone() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return stlbasic.IgnoreWith(typ.MustValue().Right())
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *hir.FuncType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e)
	})
	ret := self.analyseOptionTypeWith(node.Ret, noReturnTypeAnalyser)
	return hir.NewFuncType(ret, params...)
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

func (self *Analyser) analyseRefType(node *ast.RefType) *hir.RefType {
	return hir.NewRefType(node.Mut, self.analyseType(node.Elem))
}

func (self *Analyser) analyseSelfType(node *ast.SelfType) hir.Type {
	if !self.selfCanBeNil && self.selfType == nil {
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	} else if self.selfType != nil {
		return self.selfType
	}
	return hir.NewSelfType()
}

func (self *Analyser) analyseStructType(node *ast.StructType) *hir.StructType {
	fields := linkedhashmap.NewLinkedHashMap[string, hir.Field]()
	for _, f := range node.Fields {
		if fields.ContainKey(f.Name.Source()) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		fields.Set(f.Name.Source(), hir.Field{
			Public:  f.Public,
			Mutable: f.Mutable,
			Name:    f.Name.Source(),
			Type:    self.analyseType(f.Type),
		})
	}
	return hir.NewStructType(self.selfType, fields)
}

func (self *Analyser) analyseLambdaType(node *ast.LambdaType) *hir.LambdaType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e)
	})
	ret := self.analyseOptionTypeWith(optional.Some(node.Ret), voidTypeAnalyser, noReturnTypeAnalyser)
	return hir.NewLambdaType(ret, params...)
}

func (self *Analyser) analyseEnumType(node *ast.EnumType) *hir.EnumType {
	fields := linkedhashmap.NewLinkedHashMap[string, hir.EnumField]()
	for _, f := range node.Fields {
		if fields.ContainKey(f.Name.Source()) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		fields.Set(f.Name.Source(), hir.EnumField{
			Name: f.Name.Source(),
			Elem: stlbasic.TernaryAction(f.Elem.IsNone(), func() optional.Optional[hir.Type] {
				return optional.None[hir.Type]()
			}, func() optional.Optional[hir.Type] {
				return optional.Some(self.analyseType(f.Elem.MustValue()))
			}),
		})
	}
	return hir.NewEnumType(self.selfType, fields)
}
