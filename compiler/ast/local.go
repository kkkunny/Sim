package ast

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/optional"
)

type Local interface {
	local()
	String() string
}

type Block struct {
	Stmts []Local
}

func (b *Block) local() {}

func (b *Block) String() string {
	var buf strings.Builder
	buf.WriteString("{\n")
	for _, stmt := range b.Stmts {
		buf.WriteString("    ")
		buf.WriteString(stmt.String())
		buf.WriteString("\n")
	}
	buf.WriteString("}")
	return buf.String()
}

type Return struct {
	Value optional.Optional[Expr]
}

func (r *Return) local() {}

func (r *Return) String() string {
	if value, ok := r.Value.Value(); ok {
		return fmt.Sprintf("return %s", value.String())
	}
	return "return"
}
