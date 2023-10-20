package analyse

import (
	"math/big"

	"github.com/kkkunny/Sim/ast"
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
	switch node.Name.Source() {
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
	default:
		// TODO: 编译时异常：未知的类型
		panic("unreachable")
	}
}

func (self *Analyser) analyseFuncType(node *ast.FuncType) *FuncType {
	return &FuncType{Ret: self.analyseOptionType(node.Ret)}
}

func (self *Analyser) analyseArrayType(node *ast.ArrayType) *ArrayType {
	size, ok := big.NewInt(0).SetString(node.Size.Source(), 10)
	if !ok {
		panic("unreachable")
	} else if !size.IsUint64() {
		// TODO: 编译时异常：超出值范围
		panic("unreachable")
	}
	elem := self.analyseType(node.Elem)
	return &ArrayType{
		Size: uint(size.Uint64()),
		Elem: elem,
	}
}
