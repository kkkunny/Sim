package c

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) codegenConst(ir mir.Const)string{
	switch c := ir.(type) {
	case mir.Int:
		return self.codegenInt(c)
	case *mir.Float:
		return self.codegenFloat(c)
	case *mir.EmptyArray:
		return self.codegenEmptyArray(c)
	case *mir.EmptyStruct:
		return self.codegenEmptyStruct(c)
	case *mir.EmptyFunc:
		return self.codegenEmptyFunc(c)
	case *mir.EmptyPtr:
		return self.codegenEmptyPtr(c)
	case *mir.Array:
		return self.codegenArray(c)
	case *mir.Struct:
		return self.codegenStruct(c)
	case *mir.ConstArrayIndex:
		return self.codegenConstArrayIndex(c)
	case *mir.ConstStructIndex:
		return self.codegenConstStructIndex(c)
	case *mir.Constant:
		return self.values.Get(c)
	default:
		panic("unreachable")
	}
}

func (self *COutputer) codegenInt(ir mir.Int)string{
	s := strconv.FormatInt(ir.IntValue(), 16)
	if s[0] == '-'{
		return "0x" + s[1:]
	}else{
		return "0x" + s
	}
}

func (self *COutputer) codegenFloat(ir *mir.Float)string{
	return strconv.FormatFloat(ir.FloatValue(), 'E', -1, 64)
}

func (self *COutputer) codegenEmptyArray(ir *mir.EmptyArray)string{
	return fmt.Sprintf("(%s){}", self.codegenType(ir.Type()))
}

func (self *COutputer) codegenEmptyStruct(ir *mir.EmptyStruct)string{
	return fmt.Sprintf("(%s){}", self.codegenType(ir.Type()))
}

func (self *COutputer) codegenEmptyFunc(_ *mir.EmptyFunc)string{
	return "0"
}

func (self *COutputer) codegenEmptyPtr(_ *mir.EmptyPtr)string{
	return "0"
}

func (self *COutputer) codegenArray(ir *mir.Array)string{
	elems := lo.Map(ir.Elems(), func(item mir.Const, _ int) string {
		return self.codegenConst(item)
	})
	return fmt.Sprintf("(%s){%s}", self.codegenType(ir.Type()), strings.Join(elems, ", "))
}

func (self *COutputer) codegenStruct(ir *mir.Struct)string{
	elems := lo.Map(ir.Elems(), func(item mir.Const, _ int) string {
		return self.codegenConst(item)
	})
	return fmt.Sprintf("(%s){%s}", self.codegenType(ir.Type()), strings.Join(elems, ", "))
}

func (self *COutputer) codegenConstArrayIndex(ir *mir.ConstArrayIndex)string{
	return fmt.Sprintf("%s.data[%s]", self.codegenConst(ir.Array()), self.codegenConst(ir.Index()))
}

func (self *COutputer) codegenConstStructIndex(ir *mir.ConstStructIndex)string{
	return fmt.Sprintf("%s.e%d", self.codegenConst(ir.Struct()), ir.Index())
}
