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

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) String() string {
	var buf strings.Builder
	b.Output(&buf)
	return buf.String()
}

func (b *Builder) MoveTo(block *Block) {
	b.at = block
}

func (b *Builder) CurrentAt() (*Block, bool) {
	return b.at, b.at != nil
}

func (b *Builder) SaveFuncVarCount() int {
	n := b.funcVarCount
	b.funcVarCount = 0
	return n
}

func (b *Builder) RestoreFuncVarCount(n int) {
	b.funcVarCount = n
}

type Stmt interface {
	printWriter
	stmt()
}

type Namer interface {
	SetName(s string)
	GetName() string
}

func BuildStmt[T Stmt](b *Builder, g T) T {
	if namer, ok := any(g).(Namer); ok && stlval.Is[Local](g) && b.at != nil {
		b.funcVarCount++
		if namer.GetName() == "" {
			namer.SetName(fmt.Sprintf("_v%d", b.funcVarCount))
		}
	} else if ok && (stlval.Is[*Typedef](g) || stlval.Is[*StructTypeDef](g)) {
		b.typedefCount++
		if namer.GetName() == "" {
			namer.SetName(fmt.Sprintf("_t%d", b.typedefCount))
		}
	} else if ok && stlval.Is[Global](g) {
		b.globalVarCount++
		if namer.GetName() == "" {
			namer.SetName(fmt.Sprintf("_g%d", b.globalVarCount))
		}
	}
	if local, ok := any(g).(Local); ok && b.at != nil {
		b.at.Stmts = append(b.at.Stmts, local)
	} else {
		b.Globals = append(b.Globals, any(g).(Global))
	}
	return g
}
