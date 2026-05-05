package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type Ident struct {
	Name string

	ExternalFunc bool // 是导出的函数
}

type Context struct {
	idents    map[stmts.Ident]*Ident
	typeCache map[string]*cir.AliasType
}

func NewContext() *Context {
	return &Context{
		idents:    make(map[stmts.Ident]*Ident),
		typeCache: make(map[string]*cir.AliasType),
	}
}
