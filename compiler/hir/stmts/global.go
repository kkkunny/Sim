package stmts

import "github.com/kkkunny/Sim/compiler/hir"

type Global interface {
	hir.PrintWriter
	global()
}
