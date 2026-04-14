package cir

import (
	"fmt"
	"strings"
)

type Emitter struct {
	buf   strings.Builder
	depth int
}

func NewEmitter() *Emitter {
	return &Emitter{}
}

func (e *Emitter) Emit(program *Program) string {
	e.buf.Reset()
	for _, g := range program.Globals {
		e.emitGlobal(g)
	}
	return e.buf.String()
}

func (e *Emitter) writeIndent() {
	for i := 0; i < e.depth; i++ {
		e.buf.WriteString("    ")
	}
}

func (e *Emitter) emitType(t Type) string {
	switch t := t.(type) {
	case *VoidType:
		return "void"
	case *IntType:
		switch t.Kind {
		case IntTypeKindEnum.SChar:
			return "signed char"
		case IntTypeKindEnum.SShort:
			return "signed short"
		case IntTypeKindEnum.SInt:
			return "signed int"
		case IntTypeKindEnum.SLong:
			return "signed long"
		case IntTypeKindEnum.SLLong:
			return "signed long long"
		case IntTypeKindEnum.UChar:
			return "unsigned char"
		case IntTypeKindEnum.UShort:
			return "unsigned short"
		case IntTypeKindEnum.UInt:
			return "unsigned int"
		case IntTypeKindEnum.ULong:
			return "unsigned long"
		case IntTypeKindEnum.ULLong:
			return "unsigned long long"
		default:
			panic("unreachable")
		}
	case *FloatType:
		switch t.Kind {
		case FloatTypeKindEnum.Float:
			return "float"
		case FloatTypeKindEnum.Double:
			return "double"
		case FloatTypeKindEnum.LDouble:
			return "long double"
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

func (e *Emitter) emitGlobal(g Global) {
	switch g := g.(type) {
	case *FuncDecl:
		e.emitFunc(g)
	default:
		panic("unreachable")
	}
}

func (e *Emitter) emitFunc(fn *FuncDecl) {
	e.buf.WriteString(e.emitType(fn.ReturnType))
	e.buf.WriteString(" ")
	e.buf.WriteString(fn.Name + "(")
	for i, p := range fn.Params {
		if i > 0 {
			e.buf.WriteString(", ")
		}
		e.buf.WriteString(fmt.Sprintf("%s %s", e.emitType(p.Type), p.Name))
	}
	e.buf.WriteString(")")

	if fn.Body != nil {
		e.buf.WriteString(" ")
		e.emitBlock(fn.Body)
	} else {
		e.buf.WriteString(";")
	}
}

func (e *Emitter) emitLocal(local Local) {
	switch local := local.(type) {
	case *Block:
		e.emitBlock(local)
	case *Return:
		e.emitReturn(local)
	case Expr:
		e.emitExpr(local)
	default:
		panic("unreachable")
	}
}

func (e *Emitter) emitBlock(b *Block) {
	e.buf.WriteString("{\n")
	e.depth++
	for _, stmt := range b.Stmts {
		e.writeIndent()
		e.emitLocal(stmt)
		e.buf.WriteString("\n")
	}
	e.depth--
	e.writeIndent()
	e.buf.WriteString("}")
}

func (e *Emitter) emitReturn(r *Return) {
	e.buf.WriteString("return")
	if r.Value != nil {
		e.buf.WriteString(" ")
		e.emitExpr(r.Value)
	}
	e.buf.WriteString(";")
}

func (e *Emitter) emitExpr(expr Expr) {
	switch expr := expr.(type) {
	case *IdentExpr:
		e.buf.WriteString(expr.Name)
	case *IntegerExpr:
		e.buf.WriteString(expr.Value)
	case *BinaryExpr:
		e.buf.WriteString("(")
		e.emitExpr(expr.Left)
		e.buf.WriteString(" ")
		e.buf.WriteString(string(expr.Op))
		e.buf.WriteString(" ")
		e.emitExpr(expr.Right)
		e.buf.WriteString(")")
	default:
		panic("unreachable")
	}
}
