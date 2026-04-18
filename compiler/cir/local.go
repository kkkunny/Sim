package cir

import "github.com/kkkunny/stl/container/optional"

type Local interface {
	printWriter
	local()
}

type Block struct {
	Stmts []Local
}

func (*Block) local() {}

func (l *Block) print(p *printer) {
	p.WriteString("{")
	p.NextLine(+1)
	for i, stmt := range l.Stmts {
		p.WriteBy(stmt)
		if i < len(l.Stmts)-1 {
			p.NextLine()
		} else {
			p.NextLine(-1)
		}
	}
	p.WriteString("}")
}

type Return struct {
	Value optional.Optional[Expr]
}

func (*Return) local() {}

func (l *Return) print(p *printer) {
	p.WriteString("return")
	if value, ok := l.Value.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(value)
	}
	p.WriteString(";")
}

type VarDecl struct {
	Type  Type
	Name  string
	Value optional.Optional[Expr]
}

func (*VarDecl) local() {}

func (l *VarDecl) print(p *printer) {
	l.Type.printWithName(p, l.Name)
	if value, ok := l.Value.Value(); ok {
		p.WriteString(" = ")
		p.WriteBy(value)
	}
	p.WriteString(";")
}
