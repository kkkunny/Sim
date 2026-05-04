package codegen

import (
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type CodeGenerator struct {
	builder *cir.Builder

	idents         map[stmts.Ident]cir.Namer
	typeCache      map[string]*cir.AliasType
	captureVarsMap map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember
	currentFunc    *stmts.Func
	currentPkg     *stmts.Package
}

func New() *CodeGenerator {
	return &CodeGenerator{
		builder:        cir.NewBuilder(),
		idents:         make(map[stmts.Ident]cir.Namer),
		typeCache:      make(map[string]*cir.AliasType),
		captureVarsMap: make(map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember),
	}
}

func (c *CodeGenerator) Generate(pkg *stmts.Package) *cir.Builder {
	c.genPackage(pkg)
	return c.builder
}

func (c *CodeGenerator) genPackage(pkg *stmts.Package) {
	for _, depPkg := range pkg.Dependencies {
		c.genPackage(depPkg)
	}

	prevPkg := c.currentPkg
	c.currentPkg = pkg
	defer func() { c.currentPkg = prevPkg }()

	for _, g := range pkg.Globals {
		c.genTypeDecl(g)
	}
	for _, g := range pkg.Globals {
		c.genTypeDef(g)
	}

	for _, g := range pkg.Globals {
		c.genGlobalDecl(g)
	}
	for _, g := range pkg.Globals {
		c.genGlobalDef(g)
	}
}
