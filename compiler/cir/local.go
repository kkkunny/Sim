package cir

import (
	"fmt"

	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
)

type Local interface {
	printWriter
	local()
}

type Block struct {
	Stmts []Local

	varCount int
}

func (c *Builder) BuildBlock() *Block {
	local := &Block{}
	if c.at != nil {
		c.at.Stmts = append(c.at.Stmts, local)
	}
	return local
}

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

func (c *Builder) BuildReturn(value ...Expr) *Return {
	local := &Return{
		Value: optional.UnEmpty(stlslices.Last(value)),
	}
	c.at.Stmts = append(c.at.Stmts, local)
	return local
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
	IsGlobal bool
	Type     Type
	Name     string
	Value    optional.Optional[Expr]
}

func (c *Builder) BuildVarDecl(t Type, name string, value ...Expr) *VarDecl {
	var isGlobal bool
	if c.at != nil {
		c.at.varCount++
		if name == "" {
			name = fmt.Sprintf("_v%d", c.at.varCount)
		}
	} else {
		isGlobal = true
		c.globalVarCount++
		if name == "" {
			name = fmt.Sprintf("_g%d", c.globalVarCount)
		}
	}
	decl := &VarDecl{
		IsGlobal: isGlobal,
		Type:     t,
		Name:     name,
		Value:    optional.UnEmpty(stlslices.Last(value)),
	}
	if c.at != nil {
		c.at.Stmts = append(c.at.Stmts, decl)
	} else {
		c.Globals = append(c.Globals, decl)
	}
	return decl
}

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

func (l *VarDecl) GetName() string {
	return l.Name
}

type Include struct {
	Path string
}

func (c *Builder) BuildInclude(path string) *Include {
	i := &Include{Path: path}
	if c.at == nil {
		c.Globals = append(c.Globals, i)
	} else {
		c.at.Stmts = append(c.at.Stmts, i)
	}
	return i
}

func (*Include) local()  {}
func (*Include) global() {}

func (l *Include) print(p *printer) {
	p.WriteString("#include ")
	p.WriteString(l.Path)
}

type ExprStmt struct {
	Expr Expr
}

func (c *Builder) BuildExpr(expr Expr) *ExprStmt {
	i := &ExprStmt{Expr: expr}
	c.at.Stmts = append(c.at.Stmts, i)
	return i
}

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

func (c *Builder) BuildIf(cond Expr, body *Block, next ...either.Either[*If, *Block]) *If {
	i := &If{
		Condition: cond,
		Body:      body,
		Else:      optional.UnEmpty(stlslices.Last(next)),
	}
	c.at.Stmts = append(c.at.Stmts, i)
	return i
}

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
