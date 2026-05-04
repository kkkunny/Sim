package cir

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
)

type Local interface {
	Stmt
	local()
}

type Block struct {
	Stmts []Local
}

func NewBlock() *Block {
	return &Block{}
}

func (*Block) stmt()  {}
func (*Block) local() {}

func (l *Block) print(p *printer) {
	p.WriteString("{")
	if len(l.Stmts) > 0 {
		p.NextLine(+1)
		l.printFlat(p)
		p.NextLine(-1)
	}
	p.WriteString("}")
}

func (l *Block) printFlat(p *printer) {
	for i, stmt := range l.Stmts {
		p.WriteBy(stmt)
		if i < len(l.Stmts)-1 {
			p.NextLine()
		}
	}
}

type Return struct {
	Value optional.Optional[Expr]
}

func NewReturn(value ...Expr) *Return {
	return &Return{
		Value: optional.UnEmpty(stlslices.Last(value)),
	}
}

func (*Return) stmt()  {}
func (*Return) local() {}

func (l *Return) print(p *printer) {
	p.WriteString("return")
	if value, ok := l.Value.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(value)
	}
	p.WriteString(";")
}

type Variable struct {
	Static bool
	Type   Type
	Name   string
	Value  optional.Optional[Expr]
}

func NewVariable(t Type, name string, value ...Expr) *Variable {
	return &Variable{
		Type:  t,
		Name:  name,
		Value: optional.UnEmpty(stlslices.Last(value)),
	}
}

func (*Variable) stmt()   {}
func (*Variable) global() {}
func (*Variable) local()  {}

func (l *Variable) print(p *printer) {
	if l.Static {
		p.WriteString("static ")
	}
	l.Type.printWithName(p, l.Name)
	if value, ok := l.Value.Value(); ok {
		p.WriteString(" = ")
		p.WriteBy(value)
	}
	p.WriteString(";")
}

func (l *Variable) printDecl(p *printer) {
	if l.Static {
		p.WriteString("static ")
	}
	l.Type.printWithName(p, l.Name)
	p.WriteString(";")
}

func (l *Variable) SetName(s string) {
	l.Name = s
}

func (l *Variable) GetName() string {
	return l.Name
}

type Include struct {
	Path string
}

func NewInclude(path string) *Include {
	return &Include{Path: path}
}

func (*Include) stmt()   {}
func (*Include) global() {}
func (*Include) local()  {}

func (l *Include) print(p *printer) {
	p.WriteString("#include ")
	p.WriteString(l.Path)
}

type ExprStmt struct {
	Expr Expr
}

func NewExpr(expr Expr) *ExprStmt {
	return &ExprStmt{Expr: expr}
}

func (*ExprStmt) stmt()  {}
func (*ExprStmt) local() {}

func (l *ExprStmt) print(p *printer) {
	p.WriteBy(l.Expr)
	p.WriteString(";")
}

type If struct {
	Condition Expr
	Body      *Block
	Else      optional.Optional[either.Either[*If, *Block]]
}

func NewIf(cond Expr, body *Block, next ...either.Either[*If, *Block]) *If {
	return &If{
		Condition: cond,
		Body:      body,
		Else:      optional.UnEmpty(stlslices.Last(next)),
	}
}

func (*If) stmt()  {}
func (*If) local() {}

func (l *If) print(p *printer) {
	p.WriteString("if (")
	p.WriteBy(l.Condition)
	p.WriteString(") ")
	p.WriteBy(l.Body)

	if next, ok := l.Else.Value(); ok {
		p.WriteString(" else ")
		if elseif, ok := next.TryLeft(); ok {
			p.WriteBy(elseif)
		} else {
			p.WriteBy(next.Right())
		}
	}
}

type While struct {
	Condition Expr
	Body      *Block
}

func NewWhile(cond Expr, body *Block) *While {
	return &While{
		Condition: cond,
		Body:      body,
	}
}

func (*While) stmt()  {}
func (*While) local() {}

func (l *While) print(p *printer) {
	p.WriteString("while (")
	p.WriteBy(l.Condition)
	p.WriteString(") ")
	p.WriteBy(l.Body)
}

type For struct {
	Init      optional.Optional[*Variable]
	Condition optional.Optional[Expr]
	Action    optional.Optional[Expr]
	Body      *Block
}

func NewFor(init optional.Optional[*Variable], cond optional.Optional[Expr], action optional.Optional[Expr], body *Block) *For {
	return &For{
		Init:      init,
		Condition: cond,
		Action:    action,
		Body:      body,
	}
}

func (*For) stmt()  {}
func (*For) local() {}

func (l *For) print(p *printer) {
	p.WriteString("for (")
	if init, ok := l.Init.Value(); ok {
		p.WriteBy(init)
		p.WriteString(" ")
	} else {
		p.WriteString("; ")
	}
	if cond, ok := l.Condition.Value(); ok {
		p.WriteBy(cond)
	}
	p.WriteString("; ")
	if action, ok := l.Action.Value(); ok {
		p.WriteBy(action)
	}
	p.WriteString(") ")
	p.WriteBy(l.Body)
}

type Case struct {
	Cond Expr
	Body *Block
}

func NewCase(cond Expr, body *Block) *Case {
	return &Case{Cond: cond, Body: body}
}

type Switch struct {
	Cond    Expr
	Cases   []*Case
	Default optional.Optional[*Block]
}

func NewSwitch(cond Expr, cases ...*Case) *Switch {
	return &Switch{
		Cond:  cond,
		Cases: cases,
	}
}

func (*Switch) stmt()  {}
func (*Switch) local() {}

func (l *Switch) print(p *printer) {
	p.WriteString("switch (")
	p.WriteBy(l.Cond)
	p.WriteString(") {")
	p.NextLine()
	for _, c := range l.Cases {
		p.WriteString("case ")
		p.WriteBy(c.Cond)
		p.WriteString(":")
		if len(c.Body.Stmts) > 0 {
			p.NextLine(+1)
			c.Body.printFlat(p)
			p.NextLine(-1)
		} else {
			p.NextLine()
		}
	}
	if dc, ok := l.Default.Value(); ok {
		p.WriteString("default:")
		if len(dc.Stmts) > 0 {
			p.NextLine(+1)
			dc.printFlat(p)
			p.NextLine(-1)
		} else {
			p.NextLine()
		}
	}
	p.WriteString("}")
}
