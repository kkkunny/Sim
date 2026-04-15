package hir

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

func Print(w io.Writer, program *Program) {
	p := newPrint(w)
	p.WriteBy(program)
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
