package analyse

import (
	"math/big"

	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"

	errors "github.com/kkkunny/Sim/compiler/error"
)

type typeAnalyserFunc func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool)
type typeAnalyser = tuple.Tuple2[typeAnalyserFunc, bool]

func (self *Analyser) voidTypeAnalyser(deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if ident, ok := node.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == types.NoThing.String() {
			return types.NoThing, true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) noReturnTypeAnalyser(deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if ident, ok := node.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == types.NoReturn.String() {
			return types.NoReturn, true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) selfTypeAnalyser(deep ...bool) typeAnalyser {
	return self.selfTypeAnalyserWith(types.Self, deep...)
}

func (self *Analyser) selfTypeAnalyserWith(selfType hir.Type, deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if ident, ok := node.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == types.Self.String() {
			return selfType, true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) structTypeAnalyser(deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if st, ok := node.(*ast.StructType); ok {
			return self.analyseStructType(st, analysers...), true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) enumTypeAnalyser(deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if et, ok := node.(*ast.EnumType); ok {
			return self.analyseEnumType(et, analysers...), true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) genericParamAnalyserWith(param types.GenericParamType, deep ...bool) typeAnalyser {
	return tuple.Pack2[typeAnalyserFunc, bool](func(node ast.Type, analysers ...typeAnalyser) (hir.Type, bool) {
		if ident, ok := node.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.Name.Source() == param.String() {
			return param, true
		}
		return nil, false
	}, stlslices.Last(deep))
}

func (self *Analyser) analyseType(node ast.Type, analysers ...typeAnalyser) hir.Type {
	deepAnalysers := stlslices.Filter(analysers, func(_ int, analyser typeAnalyser) bool {
		return analyser.E2()
	})
	for _, analyser := range analysers {
		if t, ok := analyser.E1()(node, deepAnalysers...); ok {
			return t
		}
	}

	switch typeNode := node.(type) {
	case *ast.IdentType:
		return self.analyseIdentType(typeNode, deepAnalysers...)
	case *ast.FuncType:
		return self.analyseFuncType(typeNode, deepAnalysers...)
	case *ast.ArrayType:
		return self.analyseArrayType(typeNode, deepAnalysers...)
	case *ast.TupleType:
		return self.analyseTupleType(typeNode, deepAnalysers...)
	case *ast.RefType:
		return self.analyseRefType(typeNode, deepAnalysers...)
	case *ast.LambdaType:
		return self.analyseLambdaType(typeNode, deepAnalysers...)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdentType(node *ast.IdentType, analysers ...typeAnalyser) hir.Type {
	t, ok := self.tryAnalyseIdent((*ast.Ident)(node), analysers...)
	if !ok || t.IsRight() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return stlval.IgnoreWith(t.Left())
}

func (self *Analyser) analyseFuncType(node *ast.FuncType, analysers ...typeAnalyser) types.FuncType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e, analysers...)
	})
	var ret hir.Type = types.NoThing
	if retNode, ok := node.Ret.Value(); ok {
		ret = self.analyseType(retNode, append(analysers, self.noReturnTypeAnalyser())...)
	}
	return types.NewFuncType(ret, params...)
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType, analysers ...typeAnalyser) types.ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok || !size.IsUint64() {
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem, analysers...)
	return types.NewArrayType(elem, uint(size.Uint64()))
}

func (self *Analyser) analyseTupleType(node *ast.TupleType, analysers ...typeAnalyser) types.TupleType {
	elems := stlslices.Map(node.Elems, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e, analysers...)
	})
	return types.NewTupleType(elems...)
}

func (self *Analyser) analyseRefType(node *ast.RefType, analysers ...typeAnalyser) types.RefType {
	return types.NewRefType(node.Mut, self.analyseType(node.Elem, analysers...))
}

func (self *Analyser) analyseLambdaType(node *ast.LambdaType, analysers ...typeAnalyser) types.LambdaType {
	params := stlslices.Map(node.Params, func(_ int, e ast.Type) hir.Type {
		return self.analyseType(e, analysers...)
	})
	ret := self.analyseType(node.Ret, append(analysers, self.voidTypeAnalyser(), self.noReturnTypeAnalyser())...)
	return types.NewLambdaType(ret, params...)
}

func (self *Analyser) analyseStructType(node *ast.StructType, analysers ...typeAnalyser) types.StructType {
	names := set.StdHashSetWith[string]()
	return types.NewStructType(stlslices.Map(node.Fields, func(_ int, f ast.Field) *types.Field {
		name := f.Name.Source()
		if !names.Add(name) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		return types.NewField(f.Public, f.Mutable, name, self.analyseType(f.Type, analysers...))
	})...)
}

func (self *Analyser) analyseEnumType(node *ast.EnumType, analysers ...typeAnalyser) types.EnumType {
	names := set.StdHashSetWith[string]()
	return types.NewEnumType(stlslices.Map(node.Fields, func(_ int, f ast.EnumField) *types.EnumField {
		name := f.Name.Source()
		if !names.Add(name) {
			errors.ThrowIdentifierDuplicationError(f.Name.Position, f.Name)
		}
		var elem []hir.Type
		if elemNode, ok := f.Elem.Value(); ok {
			elem = append(elem, self.analyseType(elemNode, analysers...))
		}
		return types.NewEnumField(name, elem...)
	})...)
}
