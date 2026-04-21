package hir

import (
	"fmt"
	"io"
	"strings"

	stlerr "github.com/kkkunny/stl/error"
)

func Print(w io.Writer, stmt PrintWriter) {
	p := newPrint(w)
	p.WriteBy(stmt)
}

type PrintWriter interface {
	Print(p *Printer)
}

type Printer struct {
	w        io.Writer
	tabCount int
}

func newPrint(w io.Writer) *Printer {
	return &Printer{w: w}
}

func (p *Printer) WriteBy(w PrintWriter) {
	w.Print(p)
}

func (p *Printer) WriteString(s string) {
	stlerr.MustWith(p.w.Write([]byte(s)))
}

func (p *Printer) WriteFormat(f string, arg ...any) {
	p.WriteString(fmt.Sprintf(f, arg...))
}

func (p *Printer) NextLine(tab ...int) {
	p.WriteString("\n")
	for _, t := range tab {
		p.tabCount += t
	}
	p.WriteTab()
}

func (p *Printer) WriteTab() {
	p.WriteString(strings.Repeat("    ", p.tabCount))
}
