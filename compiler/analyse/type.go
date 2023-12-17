package analyse

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
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

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) hir.Type {
	t, ok := node.Value()
	if !ok {
		return hir.Empty
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) hir.Type {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.ContainKey(pkgName) {
			errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
		}
	}
	switch name := node.Name.Source(); name {
	case "isize":
		return hir.Isize
	case "i8":
		return hir.I8
	case "i16":
		return hir.I16
	case "i32":
		return hir.I32
	case "i64":
		return hir.I64
	case "i128":
		return hir.I128
	case "usize":
		return hir.Usize
	case "u8":
		return hir.U8
	case "u16":
		return hir.U16
	case "u32":
		return hir.U32
	case "u64":
		return hir.U64
	case "u128":
		return hir.U128
	case "f32":
		return hir.F32
	case "f64":
		return hir.F64
	case "bool":
		return hir.Bool
	case "str":
		return hir.Str
	case "Self":
		return hir.Self
	default:
		// 结构体
		if st, ok := self.pkgScope.GetStruct(pkgName, name); ok {
			return st
		}
		// 类型别名
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
		// FIXME: 数组最大容量
		errors.ThrowIllegalInteger(node.Position(), node.Size)
	}
	elem := self.analyseType(node.Elem)
	return &hir.ArrayType{
		Size: uint(size.Uint64()),
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

func (self *Analyser) analysePtrType(node *ast.PtrType) *hir.PtrType {
	return &hir.PtrType{Elem: self.analyseType(node.Elem)}
}

func (self *Analyser) analyseRefType(node *ast.RefType) *hir.RefType {
	return &hir.RefType{Elem: self.analyseType(node.Elem)}
}

func (self *Analyser) analyseSelfType(node *ast.SelfType) hir.Type{
	if self.selfType == nil{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}
	return self.selfType
}
