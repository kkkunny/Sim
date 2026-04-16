package cir

import "strings"

type Builder struct {
	Globals []Global
}

func NewBuilder() *Builder {
	return &Builder{}
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
