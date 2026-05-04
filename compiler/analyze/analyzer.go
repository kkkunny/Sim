package analyze

import (
	"bytes"
	"os"
	"path/filepath"

	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
)

func analyzeFile(filePath string, reporter *report.Reporter) (*stmts.Package, error) {
	data, err := stlerr.ErrorWith(os.ReadFile(filePath))
	if err != nil {
		return nil, err
	}

	lexer := lex.New(reader.NewFile(filePath, bytes.NewReader(data)))
	parser := parse.New(lexer, reporter)
	fileAst := parser.Parse()
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	analyzer := NewAnalyzer("main", filepath.Base(filePath), reporter)
	pkgHir := analyzer.Analyze(fileAst)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}
	return pkgHir, nil
}

func analyzeDir(dirPath string, reporter *report.Reporter, parent ...*Analyzer) (*stmts.Package, error) {
	pkgName := filepath.Base(dirPath)

	entries, err := stlerr.ErrorWith(os.ReadDir(dirPath))
	if err != nil {
		return nil, err
	}
	dirAsts := &ast.File{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sim" {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		data, err := stlerr.ErrorWith(os.ReadFile(filePath))
		if err != nil {
			return nil, err
		}

		lexer := lex.New(reader.NewFile(filePath, bytes.NewReader(data)))
		parser := parse.New(lexer, reporter)
		fileAst := parser.Parse()
		if reporter.HasErrors() {
			reporter.Print()
			os.Exit(1)
		}

		dirAsts.Globals = append(dirAsts.Globals, fileAst.Globals...)
	}

	var analyzer *Analyzer
	if p := stlslices.Last(parent); p != nil {
		analyzer = NewImportAnalyzer(pkgName, dirPath, p, reporter)
	} else {
		analyzer = NewAnalyzer(pkgName, dirPath, reporter)
	}
	analyzer.ir.Path = dirPath
	analyzer.pkgScopes[dirPath] = tuple.Pack2(analyzer.ir, analyzer.scope.Package())

	pkgHir := analyzer.Analyze(dirAsts)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}
	return pkgHir, nil
}

func Analyze(path string) (*stmts.Package, error) {
	info, err := stlerr.ErrorWith(os.Stat(path))
	if err != nil {
		return nil, err
	}
	reporter := report.NewReporter()
	if info.IsDir() {
		return analyzeDir(path, reporter)
	} else {
		return analyzeFile(path, reporter)
	}
}

type Analyzer struct {
	reporter *report.Reporter
	ir       *stmts.Package

	scope     scopes.Scope
	pkgScopes map[string]tuple.Tuple2[*stmts.Package, *scopes.PkgScope]

	typedefAsts map[*stmts.TypeDef]*ast.TypeDef
}

func NewAnalyzer(pkgName string, pkgPath string, reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter:    reporter,
		ir:          &stmts.Package{Name: pkgName, Path: pkgPath},
		scope:       scopes.NewPkgScope(pkgName),
		pkgScopes:   make(map[string]tuple.Tuple2[*stmts.Package, *scopes.PkgScope]),
		typedefAsts: make(map[*stmts.TypeDef]*ast.TypeDef),
	}
}

func NewImportAnalyzer(pkgName string, pkgPath string, parent *Analyzer, reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter:    reporter,
		ir:          &stmts.Package{Name: pkgName, Path: pkgPath},
		scope:       scopes.NewPkgScope(pkgName),
		pkgScopes:   parent.pkgScopes,
		typedefAsts: make(map[*stmts.TypeDef]*ast.TypeDef),
	}
}

func (a *Analyzer) Scope() scopes.Scope {
	return a.scope
}

func (a *Analyzer) Analyze(program *ast.File) *stmts.Package {
	for _, g := range program.Globals {
		importAst, ok := g.(*ast.Import)
		if !ok {
			continue
		}
		err := a.analyzeImport(importAst)
		if err != nil {
			panic(err)
		}
	}

	for _, t := range program.Globals {
		decl := a.analyzeTypeDecl(t)
		if decl == nil {
			continue
		}
		a.ir.Globals = append(a.ir.Globals, decl)
	}
	for _, t := range program.Globals {
		a.analyzeTypeDef(t)
	}

	for _, v := range program.Globals {
		a.analyzeGlobalValueDecl(v)
	}

	for _, v := range program.Globals {
		def := a.analyzeGlobalValueDef(v)
		if def == nil {
			continue
		}
		a.ir.Globals = append(a.ir.Globals, def)
	}
	return a.ir
}
