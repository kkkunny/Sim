package cir

import (
	"fmt"
	"strings"

	stlval "github.com/kkkunny/stl/value"
)

type Builder struct {
	Globals []Global

	at *Block

	typedefCount   int
	globalVarCount int
	funcVarCount   int
}

func (b *Builder) SaveFuncVarCount() int {
	n := b.funcVarCount
	b.funcVarCount = 0
	return n
}

func (b *Builder) RestoreFuncVarCount(n int) {
	b.funcVarCount = n
}

func NewBuilder() *Builder {
	b := &Builder{}
	BuildStmt(b, NewInclude("\"buildin.c\""))
	return b
}

func (b *Builder) print(p *printer) {
	for i, g := range b.Globals {
		p.WriteBy(g)
		if i != len(b.Globals)-1 {
			p.WriteString("\n")
		}
	}
}

func (c *Builder) String() string {
	var buf strings.Builder
	c.Output(&buf)
	return buf.String()
}

func (c *Builder) MoveTo(b *Block) {
	c.at = b
}

func (c *Builder) CurrentAt() (*Block, bool) {
	return c.at, c.at != nil
}

type Stmt interface {
	printWriter
	stmt()
}

func BuildStmt[T Stmt](b *Builder, g T) T {
	if namer, ok := any(g).(Namer); ok && stlval.Is[Global](g) && stlval.Is[Local](g) {
		if b.at != nil {
			b.funcVarCount++
			if namer.GetName() == "" {
				namer.SetName(fmt.Sprintf("_v%d", b.funcVarCount))
			}
		} else {
			if v, ok := namer.(*VarDecl); ok {
				v.IsGlobal = true
			}
			b.globalVarCount++
			if namer.GetName() == "" {
				namer.SetName(fmt.Sprintf("_g%d", b.globalVarCount))
			}
		}
	} else if ok && stlval.Is[*Typedef](g) {
		b.typedefCount++
		if namer.GetName() == "" {
			namer.SetName(fmt.Sprintf("_t%d", b.typedefCount))
		}
	}
	if local, ok := any(g).(Local); ok {
		if b.at != nil {
			b.at.Stmts = append(b.at.Stmts, local)
		}
	} else {
		b.Globals = append(b.Globals, any(g).(Global))
	}
	return g
}
