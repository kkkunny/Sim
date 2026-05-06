package analyze

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/kkkunny/stl/container/bimap"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
)

func analyzeFile(filePath string, reporter *report.Reporter) (*globals.Package, error) {
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

	analyzer := NewAnalyzer("main", filepath.Dir(filePath), reporter)
	pkgHir := analyzer.Analyze(fileAst)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}
	return pkgHir, nil
}

func analyzeDir(dirPath string, reporter *report.Reporter, parent ...*Analyzer) (*globals.Package, error) {
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
	analyzer.pkgScopes[dirPath] = tuple.Pack2(analyzer.ir, analyzer.scope.Root())

	pkgHir := analyzer.Analyze(dirAsts)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}
	return pkgHir, nil
}

func Analyze(path string) (*globals.Package, error) {
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
	ir       *globals.Package

	scope     scopes.Scope
	pkgScopes map[string]tuple.Tuple2[*globals.Package, *scopes.PkgScope]

	typeDef2Ast  bimap.BiMap[*globals.TypeDef, *ast.TypeDef]
	typeName2Def map[string]*globals.TypeDef

	letDef2Ast  map[*locals.Let]*ast.Let
	letDefStack set.Set[*locals.Let]
}

func NewAnalyzer(pkgName string, pkgPath string, reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter: reporter,
		ir:       &globals.Package{Name: pkgName, Path: pkgPath},

		scope:     scopes.NewPkgScope(pkgName),
		pkgScopes: make(map[string]tuple.Tuple2[*globals.Package, *scopes.PkgScope]),

		typeDef2Ast:  bimap.StdWith[*globals.TypeDef, *ast.TypeDef](),
		typeName2Def: make(map[string]*globals.TypeDef),

		letDef2Ast:  make(map[*locals.Let]*ast.Let),
		letDefStack: set.StdHashSetWith[*locals.Let](),
	}
}

func NewImportAnalyzer(pkgName string, pkgPath string, parent *Analyzer, reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter: reporter,
		ir:       &globals.Package{Name: pkgName, Path: pkgPath},

		scope:     scopes.NewPkgScope(pkgName),
		pkgScopes: parent.pkgScopes,

		typeDef2Ast:  bimap.StdWith[*globals.TypeDef, *ast.TypeDef](),
		typeName2Def: make(map[string]*globals.TypeDef),

		letDef2Ast:  make(map[*locals.Let]*ast.Let),
		letDefStack: set.StdHashSetWith[*locals.Let](),
	}
}

func (a *Analyzer) Scope() scopes.Scope {
	return a.scope
}

func (a *Analyzer) Analyze(program *ast.File) *globals.Package {
	a.analyzePackageImport(program)
	a.analyzeGlobalType(program)
	a.analyzeGlobalValue(program)
	return a.ir
}

func (a *Analyzer) analyzePackageImport(program *ast.File) {
	err := a.importBuildin()
	if err != nil {
		panic(err)
	}
	for _, g := range program.Globals {
		importAst, ok := g.(*ast.Import)
		if !ok {
			continue
		}
		err = a.analyzeImport(importAst)
		if err != nil {
			panic(err)
		}
	}
}

func (a *Analyzer) analyzeGlobalType(program *ast.File) {
	for _, g := range program.Globals {
		decl := a.analyzeTypePreDecl(g)
		if decl == nil {
			continue
		}
		a.ir.Globals = append(a.ir.Globals, decl)
	}
	for _, g := range program.Globals {
		a.analyzeTypeDecl(g)
	}
	for _, g := range program.Globals {
		a.analyzeTypeDef(g)
	}
	// 检查循环引用
	for _, g := range program.Globals {
		t, ok := g.(*ast.TypeDef)
		if !ok {
			continue
		}
		ct, _ := a.scope.LookupType(t.Name.OriginText)
		if types.CheckRecursion(ct) {
			a.reporter.Fatalf(
				t.Name.Position,
				report.Errors.CircularReference,
			)
		}
	}
}

func (a *Analyzer) analyzeGlobalValue(program *ast.File) {
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
}
