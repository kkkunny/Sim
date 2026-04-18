package cir

import (
	"strings"

	stlval "github.com/kkkunny/stl/value"
)

type Builder struct {
	Globals []Global

	typedefCount int
	varCount     int
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) print(p *printer) {
	for i, g := range b.Globals {
		if !stlval.Is[*Typedef](g) {
			continue
		}
		p.WriteBy(g)
		if i != len(b.Globals)-1 {
			p.WriteString("\n")
		}
	}
	for i, g := range b.Globals {
		if stlval.Is[*Typedef](g) {
			continue
		}
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
