package analyse

import (
	"math/big"

	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/oldhir"

	errors "github.com/kkkunny/Sim/compiler/error"
)

func (self *Analyser) analyseType(node ast.Type) oldhir.Type {
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

var voidTypeAnalyser = func(node optional.Optional[ast.Type]) optional.Optional[oldhir.Type] {
	typeNode, ok := node.Value()
	if !ok {
		return optional.None[oldhir.Type]()
	}
	if ident, ok := typeNode.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == oldhir.NoThing.String() {
		return optional.Some[oldhir.Type](oldhir.NoThing)
	}
	return optional.None[oldhir.Type]()
}

var noReturnTypeAnalyser = func(node optional.Optional[ast.Type]) optional.Optional[oldhir.Type] {
	typeNode, ok := node.Value()
	if !ok {
		return optional.None[oldhir.Type]()
	}
	if ident, ok := typeNode.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == oldhir.NoReturn.String() {
		return optional.Some[oldhir.Type](oldhir.NoReturn)
	}
	return optional.None[oldhir.Type]()
}

func (self *Analyser) analyseOptionType(node optional.Optional[ast.Type]) oldhir.Type {
	t, ok := node.Value()
	if !ok {
		return oldhir.NoThing
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseOptionTypeWith(node optional.Optional[ast.Type], analysers ...func(node optional.Optional[ast.Type]) optional.Optional[oldhir.Type]) oldhir.Type {
	for _, analyser := range analysers {
		if t, ok := analyser(node).Value(); ok {
			return t
		}
	}
	return self.analyseOptionType(node)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) oldhir.Type {
	typ := self.analyseIdent((*ast.Ident)(node), false)
	if typ.IsNone() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return stlval.IgnoreWith(typ.MustValue().Right())
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *oldhir.FuncType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) oldhir.Type {
		return self.analyseType(e)
	})
	ret := self.analyseOptionTypeWith(node.Ret, noReturnTypeAnalyser)
	return oldhir.NewFuncType(ret, params...)
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *oldhir.ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return oldhir.NewArrayType(size.Uint64(), elem)
}

func (self *Analyser) analyseTupleType(node *ast.TupleType) *oldhir.TupleType {
	elems := stlslices.Map(node.Elems, func(_ int, e ast.Type) oldhir.Type {
		return self.analyseType(e)
	})
	return oldhir.NewTupleType(elems...)
}

func (self *Analyser) analyseRefType(node *ast.RefType) *oldhir.RefType {
	return oldhir.NewRefType(node.Mut, self.analyseType(node.Elem))
}

func (self *Analyser) analyseSelfType(node *ast.SelfType) oldhir.Type {
	if !self.selfCanBeNil && self.selfType == nil {
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	} else if self.selfType != nil {
		return self.selfType
	}
	return oldhir.NewSelfType()
}

func (self *Analyser) analyseStructType(node *ast.StructType) *oldhir.StructType {
	fields := linkedhashmap.StdWith[string, oldhir.Field]()
	for _, f := range node.Fields {
		if fields.Contain(f.Name.Source()) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		fields.Set(f.Name.Source(), oldhir.Field{
			Public:  f.Public,
			Mutable: f.Mutable,
			Name:    f.Name.Source(),
			Type:    self.analyseType(f.Type),
		})
	}
	return oldhir.NewStructType(self.selfType, fields)
}

func (self *Analyser) analyseLambdaType(node *ast.LambdaType) *oldhir.LambdaType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) oldhir.Type {
		return self.analyseType(e)
	})
	ret := self.analyseOptionTypeWith(optional.Some(node.Ret), voidTypeAnalyser, noReturnTypeAnalyser)
	return oldhir.NewLambdaType(ret, params...)
}

func (self *Analyser) analyseEnumType(node *ast.EnumType) *oldhir.EnumType {
	fields := linkedhashmap.StdWith[string, oldhir.EnumField]()
	for _, f := range node.Fields {
		if fields.Contain(f.Name.Source()) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		fields.Set(f.Name.Source(), oldhir.EnumField{
			Name: f.Name.Source(),
			Elem: stlval.TernaryAction(f.Elem.IsNone(), func() optional.Optional[oldhir.Type] {
				return optional.None[oldhir.Type]()
			}, func() optional.Optional[oldhir.Type] {
				return optional.Some(self.analyseType(f.Elem.MustValue()))
			}),
		})
	}
	return oldhir.NewEnumType(self.selfType, fields)
}
