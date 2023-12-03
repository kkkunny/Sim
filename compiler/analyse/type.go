package analyse

import (
	"math/big"

	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseType(node ast.Type) mean.Type {
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
	case *ast.PtrType:
		return self.analysePtrType(typeNode)
	case *ast.RefType:
		return self.analyseRefType(typeNode)
	case *ast.SelfType:
		return self.analyseSelfType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) mean.Type {
	t, ok := node.Value()
	if !ok {
		return mean.Empty
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) mean.Type {
	switch name := node.Name.Source(); name {
	case "isize":
		return mean.Isize
	case "i8":
		return mean.I8
	case "i16":
		return mean.I16
	case "i32":
		return mean.I32
	case "i64":
		return mean.I64
	case "i128":
		return mean.I128
	case "usize":
		return mean.Usize
	case "u8":
		return mean.U8
	case "u16":
		return mean.U16
	case "u32":
		return mean.U32
	case "u64":
		return mean.U64
	case "u128":
		return mean.U128
	case "f32":
		return mean.F32
	case "f64":
		return mean.F64
	case "bool":
		return mean.Bool
	case "str":
		return mean.Str
	default:
		var pkgName string
		if pkgToken, ok := node.Pkg.Value(); ok {
			pkgName = pkgToken.Source()
			if !self.pkgScope.externs.ContainKey(pkgName) {
				errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
			}
		}
		if st, ok := self.pkgScope.GetStruct(pkgName, name); ok {
			return st
		}
		if typeAlias, ok := self.pkgScope.GetTypeAlias(pkgName, name); ok {
			if target, ok := typeAlias.Right(); ok{
				return target
			}
			return self.defTypeAlias(name)
		}
		errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
		return nil
	}
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *mean.FuncType {
	params := lo.Map(node.Params, func(item ast.Type, index int) mean.Type {
		return self.analyseType(item)
	})
	return &mean.FuncType{
		Ret:    self.analyseOptionType(node.Ret),
		Params: params,
	}
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *mean.ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		// FIXME: 数组最大容量
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return &mean.ArrayType{
		Size: uint(size.Uint64()),
		Elem: elem,
	}
}

func (self *Analyser) analyseTupleType(node *ast.TupleType) *mean.TupleType {
	elems := lo.Map(node.Elems, func(item ast.Type, index int) mean.Type {
		return self.analyseType(item)
	})
	return &mean.TupleType{Elems: elems}
}

func (self *Analyser) analyseUnionType(node *ast.UnionType) *mean.UnionType {
	elems := iterator.Map[ast.Type, pair.Pair[string, mean.Type], linkedhashmap.LinkedHashMap[string, mean.Type]](node.Elems, func(v ast.Type) pair.Pair[string, mean.Type] {
		et := self.analyseType(v)
		return pair.NewPair(et.String(), et)
	})
	return &mean.UnionType{Elems: elems}
}

func (self *Analyser) analysePtrType(node *ast.PtrType) *mean.PtrType {
	return &mean.PtrType{Elem: self.analyseType(node.Elem)}
}

func (self *Analyser) analyseRefType(node *ast.RefType) *mean.RefType {
	return &mean.RefType{Elem: self.analyseType(node.Elem)}
}

func (self *Analyser) analyseSelfType(node *ast.SelfType)*mean.StructType{
	if self.selfType == nil{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}
	return self.selfType
}
