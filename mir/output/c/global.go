package c

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) codegenDeclType(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.NamedStruct:
		name := fmt.Sprintf("_sim_type_%d", self.typedefs.Length())
		self.typedefs.Set(fmt.Sprintf("%p", global), pair.NewPair(name, ""))
		self.types.Set(global, name)
		self.structDeclBuffer.WriteString(fmt.Sprintf("struct %s;\n", name))
	case *mir.GlobalVariable, *mir.Constant, *mir.Function:
	default:
		panic("unreachable")
	}
}

func (self *COutputer) codegenDefType(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.NamedStruct:
		name := self.types.Get(global)
		elems := stlslices.Map(global.Elems(), func(_ int, e mir.Type) string {
			return self.codegenType(e)
		})
		var buf strings.Builder
		buf.WriteString("struct ")
		buf.WriteString(name)
		buf.WriteString("{\n")
		for i, e := range elems {
			buf.WriteByte('\t')
			buf.WriteString(e)
			buf.WriteByte(' ')
			buf.WriteString(fmt.Sprintf("e%d", i))
			buf.WriteString(";\n")
		}
		buf.WriteByte('}')
		self.typedefs.Set(fmt.Sprintf("%p", global), pair.NewPair(name, buf.String()))
	case *mir.GlobalVariable, *mir.Constant, *mir.Function:
	default:
		panic("unreachable")
	}
}

func (self *COutputer) getValueDecl(ir mir.Global, name string) string {
	var buf strings.Builder
	switch global := ir.(type) {
	case *mir.GlobalVariable:
		if global.RealName() != "" {
			buf.WriteString("extern ")
		} else {
			buf.WriteString("static ")
		}
		buf.WriteString(fmt.Sprintf("%s %s", self.codegenType(global.ValueType()), name))
	case *mir.Constant:
		if global.RealName() != "" {
			buf.WriteString("extern ")
		} else {
			buf.WriteString("static ")
		}
		buf.WriteString(fmt.Sprintf("const %s %s", self.codegenType(global.ValueType()), name))
	case *mir.Function:
		ftIr := global.Type().(mir.FuncType)
		ret := self.codegenType(ftIr.Ret())
		params := stlslices.Map(global.Params(), func(i int, e *mir.Param) string {
			return fmt.Sprintf("%s %s", self.codegenType(e.ValueType()), fmt.Sprintf("_p%d", i))
		})
		if global.RealName() != "" {
			buf.WriteString("extern ")
		} else {
			buf.WriteString("static ")
		}
		buf.WriteString(fmt.Sprintf("%s %s(%s)", ret, name, strings.Join(params, ", ")))
	default:
		panic("unreachable")
	}
	return buf.String()
}

func (self *COutputer) codegenDeclValue(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.NamedStruct:
	case *mir.GlobalVariable:
		name := stlbasic.TernaryAction(global.RealName() != "", func() string {
			return global.RealName()
		}, func() string {
			return fmt.Sprintf("_sim_var_%d", self.values.Length())
		})
		self.values.Set(global, name)
		self.varDeclBuffer.WriteString(self.getValueDecl(global, name))
		self.varDeclBuffer.WriteString(";\n")
	case *mir.Constant:
		name := stlbasic.TernaryAction(global.RealName() != "", func() string {
			return global.RealName()
		}, func() string {
			return fmt.Sprintf("_sim_const_%d", self.values.Length())
		})
		self.values.Set(global, name)
		self.varDeclBuffer.WriteString(self.getValueDecl(global, name))
		self.varDeclBuffer.WriteString(";\n")
	case *mir.Function:
		name := stlbasic.TernaryAction(global.RealName() != "", func() string {
			return global.RealName()
		}, func() string {
			return fmt.Sprintf("_sim_func_%d", self.values.Length())
		})
		self.values.Set(global, name)
		self.varDeclBuffer.WriteString(self.getValueDecl(global, name))
		self.varDeclBuffer.WriteString(";\n")
	default:
		panic("unreachable")
	}
}

func (self *COutputer) codegenDefValue(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.NamedStruct:
	case *mir.GlobalVariable:
		if value, ok := global.Value(); ok {
			self.varDefBuffer.WriteString(self.getValueDecl(global, self.values.Get(global)))
			self.varDefBuffer.WriteString(" = ")
			self.varDefBuffer.WriteString(self.codegenConst(value))
			self.varDefBuffer.WriteString(";\n")
		}
	case *mir.Constant:
		self.varDefBuffer.WriteString(self.getValueDecl(global, self.values.Get(global)))
		self.varDefBuffer.WriteString(" = ")
		self.varDefBuffer.WriteString(self.codegenConst(global.Value()))
		self.varDefBuffer.WriteString(";\n")
	case *mir.Function:
		if global.Blocks().Len() == 0 {
			return
		}
		self.varDefBuffer.WriteString(self.getValueDecl(global, self.values.Get(global)))
		for i, p := range global.Params() {
			self.values.Set(p, fmt.Sprintf("_p%d", i))
		}
		self.varDefBuffer.WriteString("{\n")
		self.localVarCount = 0
		for cursor := global.Blocks().Front(); cursor != nil; cursor = cursor.Next() {
			bn := fmt.Sprintf("_b%d", self.blocks.Length())
			self.blocks.Set(cursor.Value, bn)
		}
		for cursor := global.Blocks().Front(); cursor != nil; cursor = cursor.Next() {
			self.codegenBlock(cursor.Value)
		}
		self.varDefBuffer.WriteString("}\n")
	default:
		panic("unreachable")
	}
}

func (self *COutputer) include(path string) {
	if self.includes.Contain(path) {
		return
	}
	self.includes.Add(path)
}
