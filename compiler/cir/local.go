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
		for i, stmt := range l.Stmts {
			p.WriteBy(stmt)
			if i < len(l.Stmts)-1 {
				p.NextLine()
			} else {
				p.NextLine(-1)
			}
		}
	}
	p.WriteString("}")
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

type VarDecl struct {
	IsGlobal bool
	Type     Type
	Name     string
	Value    optional.Optional[Expr]
}

func NewVarDecl(t Type, name string, value ...Expr) *VarDecl {
	return &VarDecl{
		Type:  t,
		Name:  name,
		Value: optional.UnEmpty(stlslices.Last(value)),
	}
}

func (*VarDecl) stmt()   {}
func (*VarDecl) global() {}
func (*VarDecl) local()  {}

func (l *VarDecl) print(p *printer) {
	if l.IsGlobal {
		p.WriteString("static ")
	}
	l.Type.printWithName(p, l.Name)
	if value, ok := l.Value.Value(); ok {
		p.WriteString(" = ")
		p.WriteBy(value)
	}
	p.WriteString(";")
}

func (l *VarDecl) SetName(s string) {
	l.Name = s
}

func (l *VarDecl) GetName() string {
	return l.Name
}

type Include struct {
	Path string
}

func NewInclude(path string) *Include {
	return &Include{Path: path}
}

func (*Include) stmt()   {}
func (*Include) local()  {}
func (*Include) global() {}

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

func (*ExprStmt) stmt()   {}
func (*ExprStmt) local()  {}
func (*ExprStmt) global() {}

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

func (*If) stmt()   {}
func (*If) local()  {}
func (*If) global() {}

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

func (*While) stmt()   {}
func (*While) local()  {}
func (*While) global() {}

func (l *While) print(p *printer) {
	p.WriteString("while (")
	p.WriteBy(l.Condition)
	p.WriteString(") ")
	p.WriteBy(l.Body)
}

type For struct {
	Init      optional.Optional[*VarDecl]
	Condition optional.Optional[Expr]
	Action    optional.Optional[Expr]
	Body      *Block
}

func NewFor(init optional.Optional[*VarDecl], cond optional.Optional[Expr], action optional.Optional[Expr], body *Block) *For {
	return &For{
		Init:      init,
		Condition: cond,
		Action:    action,
		Body:      body,
	}
}

func (*For) stmt()   {}
func (*For) local()  {}
func (*For) global() {}

func (l *For) print(p *printer) {
	p.WriteString("for (")
	if init, ok := l.Init.Value(); ok {
		p.WriteBy(init)
	}
	p.WriteString("; ")
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
