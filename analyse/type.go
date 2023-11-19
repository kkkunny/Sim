package analyse

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseType(node ast.Type) Type {
	switch typeNode := node.(type) {
	case *ast.IdentType:
		return self.analyseIdentType(typeNode)
	case *ast.FuncType:
		return self.analyseFuncType(typeNode)
	case *ast.ArrayType:
		return self.analyseArrayType(typeNode)
	case *ast.TupleType:
		return self.analyseTupleType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) Type {
	t, ok := node.Value()
	if !ok {
		return Empty
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) Type {
	switch name := node.Name.Source(); name {
	case "isize":
		return Isize
	case "i8":
		return I8
	case "i16":
		return I16
	case "i32":
		return I32
	case "i64":
		return I64
	case "i128":
		return I128
	case "usize":
		return Usize
	case "u8":
		return U8
	case "u16":
		return U16
	case "u32":
		return U32
	case "u64":
		return U64
	case "u128":
		return U128
	case "f32":
		return F32
	case "f64":
		return F64
	case "bool":
		return Bool
	case "str":
		return Str
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
		} else {
			errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
		}
		return nil
	}
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *FuncType {
	params := lo.Map(node.Params, func(item ast.Type, index int) Type {
		return self.analyseType(item)
	})
	return &FuncType{
		Ret:    self.analyseOptionType(node.Ret),
		Params: params,
	}
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		// FIXME: 数组最大容量
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return &ArrayType{
		Size: uint(size.Uint64()),
		Elem: elem,
	}
}

func (self *Analyser) analyseTupleType(node *ast.TupleType) *TupleType {
	elems := lo.Map(node.Elems, func(item ast.Type, index int) Type {
		return self.analyseType(item)
	})
	return &TupleType{Elems: elems}
}
