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
		return self.analyseIdentType(typeNode)
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

func (self *Analyser) tryAnalyseIdentType(node *ast.IdentType) (hir.Type, bool) {
	name := node.Name.Source()

	// 仅buildin包类型
	if self.pkg.IsBuildIn() {
		switch name {
		case "__buildin_isize":
			return types.Isize, true
		case "__buildin_i8":
			return types.I8, true
		case "__buildin_i16":
			return types.I16, true
		case "__buildin_i32":
			return types.I32, true
		case "__buildin_i64":
			return types.I64, true
		case "__buildin_usize":
			return types.Usize, true
		case "__buildin_u8":
			return types.U8, true
		case "__buildin_u16":
			return types.U16, true
		case "__buildin_u32":
			return types.U32, true
		case "__buildin_u64":
			return types.U64, true
		case "__buildin_f16":
			return types.F16, true
		case "__buildin_f32":
			return types.F32, true
		case "__buildin_f64":
			return types.F64, true
		case "__buildin_f128":
			return types.F128, true
		case "__buildin_bool":
			return types.Bool, true
		case "__buildin_str":
			return types.Str, true
		}
	}

	scope := self.scope
	if pkgToken, ok := node.Pkg.Value(); ok {
		scope, ok = self.pkg.GetExternPackage(pkgToken.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	// 自定义类型
	obj, ok := scope.GetIdent(name, true)
	if ok && stlval.Is[hir.Type](obj) {
		return obj.(hir.Type), true
	}
	return nil, false
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) hir.Type {
	t, ok := self.tryAnalyseIdentType(node)
	if !ok {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return t
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
