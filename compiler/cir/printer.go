package cir

import (
	"fmt"
	"io"
	"strings"

	stlerr "github.com/kkkunny/stl/error"
)

type printWriter interface {
	print(p *printer)
}

type printer struct {
	w        io.Writer
	tabCount int
}

func newPrint(w io.Writer) *printer {
	return &printer{w: w}
}

func (b *Builder) Output(w io.Writer) {
	b.OutputHeader(w)
	b.OutputSource(w)
}

func (b *Builder) OutputHeader(w io.Writer) {
	p := newPrint(w)
	for i, g := range b.Globals {
		switch g := g.(type) {
		case *Func:
			if g.Body.IsSome() {
				continue
			}
			g.printDecl(p)
		case *Variable:
			if g.Value.IsSome() {
				continue
			}
			g.printDecl(p)
		default:
			p.WriteBy(g)
		}
		if i != len(b.Globals)-1 {
			p.WriteString("\n")
		}
	}
}

func (b *Builder) OutputSource(w io.Writer) {
	p := newPrint(w)
	for i, g := range b.Globals {
		switch g := g.(type) {
		case *Func:
			if g.Body.IsNone() {
				continue
			}
		case *Variable:
			if g.Value.IsNone() {
				continue
			}
		default:
			continue
		}
		g.print(p)
		if i != len(b.Globals)-1 {
			p.WriteString("\n")
		}
	}
}

func (p *printer) WriteBy(w printWriter) {
	w.print(p)
}

func (p *printer) WriteString(s string) {
	stlerr.MustWith(p.w.Write([]byte(s)))
}

func (p *printer) WriteFormat(f string, arg ...any) {
	p.WriteString(fmt.Sprintf(f, arg...))
}

func (p *printer) NextLine(tab ...int) {
	p.WriteString("\n")
	for _, t := range tab {
		p.tabCount += t
	}
	p.WriteTab()
}

func (p *printer) WriteTab() {
	p.WriteString(strings.Repeat("    ", p.tabCount))
}
