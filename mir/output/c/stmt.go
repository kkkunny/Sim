package c

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) newLocalName()string{
	self.localVarCount++
	return fmt.Sprintf("_v%d", self.localVarCount)
}

func (self *COutputer) codegenStmt(ir mir.Stmt){
	switch stmt := ir.(type) {
	case *mir.Store:
		self.varDefBuffer.WriteString(fmt.Sprintf("%s = %s", self.codegenValue(stmt.To()), self.codegenValue(stmt.From())))
	case *mir.Return:
		self.varDefBuffer.WriteString("return")
		if value, ok := stmt.Value(); ok{
			self.varDefBuffer.WriteByte(' ')
			self.varDefBuffer.WriteString(self.codegenValue(value))
		}
	case *mir.Unreachable:
		self.varDefBuffer.WriteString("# unreachable")
	case *mir.UnCondJump:
		self.varDefBuffer.WriteString(fmt.Sprintf("goto %s", self.blocks.Get(stmt.To())))
	case *mir.CondJump:
		self.varDefBuffer.WriteString(fmt.Sprintf("if (%s) goto %s; else goto %s;", self.codegenValue(stmt.Cond()), self.blocks.Get(stmt.TrueBlock()), self.blocks.Get(stmt.FalseBlock())))
	case mir.StmtValue:
		self.codegenStmtValue(stmt)
	default:
		panic("unreachable")
	}
}

func (self *COutputer) codegenStmtValue(ir mir.StmtValue){
	var name string
	defer func() {
		self.values.Set(ir, name)
	}()

	tname := self.codegenType(ir.Type())
	switch stmt := ir.(type) {
	case *mir.AllocFromStack:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s", self.codegenType(stmt.ElemType()), name))
	case *mir.AllocFromHeap:
		self.include("<stdlib.h>")
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = calloc(1, sizeof(%s))", self.codegenType(stmt.ElemType()), name, stmt.Type()))
	case *mir.Load:
		if stlbasic.Is[*mir.AllocFromStack](stmt.From()) || stlbasic.Is[*mir.Param](stmt.From()){
			name = self.codegenValue(stmt.From())
		}else{
			name = self.newLocalName()
			self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = *%s", tname, name, self.codegenValue(stmt.From())))
		}
	case *mir.And:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s & %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Or:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s | %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Xor:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s ^ %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Shl:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s << %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Shr:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s >> %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Not:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = ~%s", tname, name, self.codegenValue(stmt.Value())))
	case *mir.Add:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s + %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Sub:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s - %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Mul:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s * %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Div:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s / %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Rem:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s % %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Cmp:
		var opera string
		switch stmt.Kind() {
		case mir.CmpKindEQ:
			opera = "=="
		case mir.CmpKindNE:
			opera = "!="
		case mir.CmpKindLT:
			opera = "<"
		case mir.CmpKindLE:
			opera = "<="
		case mir.CmpKindGT:
			opera = ">"
		case mir.CmpKindGE:
			opera = ">="
		default:
			panic("unreachable")
		}
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s %s %s", tname, name, self.codegenValue(stmt.Left()), opera, self.codegenValue(stmt.Right())))
	case *mir.PtrEqual:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s == %s", tname, name, self.codegenValue(stmt.Left()), self.codegenValue(stmt.Right())))
	case *mir.Call:
		args := stlslices.Map(stmt.Args(), func(_ int, e mir.Value) string {
			return self.codegenValue(e)
		})
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s(%s)", tname, name, self.codegenValue(stmt.Func()), strings.Join(args, ", ")))
	case *mir.NumberCovert:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = (%s)%s", tname, name, tname, self.codegenValue(stmt.From())))
	case *mir.PtrToUint:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = (%s)%s", tname, name, tname, self.codegenValue(stmt.From())))
	case *mir.UintToPtr:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = (%s)%s", tname, name, tname, self.codegenValue(stmt.From())))
	case *mir.PtrToPtr:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = (%s)%s", tname, name, tname, self.codegenValue(stmt.From())))
	case *mir.ArrayIndex:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s.data[%s]", tname, name, self.codegenValue(stmt.Array()), self.codegenValue(stmt.Index())))
	case *mir.StructIndex:
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = %s.e%d", tname, name, self.codegenValue(stmt.Struct()), stmt.Index()))
	case *mir.Phi:
		// TODO: phi
	case *mir.PackArray:
		elems := stlslices.Map(stmt.Elems(), func(_ int, e mir.Value) string {
			return self.codegenValue(e)
		})
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = {%s}", tname, name, strings.Join(elems, ", ")))
	case *mir.PackStruct:
		elems := stlslices.Map(stmt.Elems(), func(_ int, e mir.Value) string {
			return self.codegenValue(e)
		})
		name = self.newLocalName()
		self.varDefBuffer.WriteString(fmt.Sprintf("%s %s = {%s}", tname, name, strings.Join(elems, ", ")))
	default:
		panic("unreachable")
	}
}
