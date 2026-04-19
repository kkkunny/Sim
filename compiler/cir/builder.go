package cir

import (
	"strings"
)

type Builder struct {
	Globals []Global

	at *Block

	typedefCount   int
	globalVarCount int
}

func NewBuilder() *Builder {
	b := &Builder{}
	b.BuildInclude("\"buildin.c\"")
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
