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
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
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
		return nil, report.ErrReported
	}

	analyzer := NewAnalyzer("main", filepath.Dir(filePath), reporter)
	return analyzer.Analyze(fileAst)
}

func analyzeDir(dirPath string, reporter *report.Reporter, parent ...*Analyzer) (*globals.Package, error) {
	pkgName := filepath.Base(dirPath)

	entries, err := stlerr.ErrorWith(os.ReadDir(dirPath))
	if err != nil {
		return nil, err
	}
	dirAsts := &ast.File{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != config.SourceCodeFileExtName {
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

		dirAsts.Globals = append(dirAsts.Globals, fileAst.Globals...)
	}
	if reporter.HasErrors() {
		return nil, report.ErrReported
	}

	var analyzer *Analyzer
	if p := stlslices.Last(parent); p != nil {
		analyzer = NewImportAnalyzer(pkgName, dirPath, p, reporter)
	} else {
		analyzer = NewAnalyzer(pkgName, dirPath, reporter)
	}
	analyzer.ir.Path = dirPath
	analyzer.pkgScopes[dirPath] = tuple.Pack2(analyzer.ir, analyzer.scope.Root())

	return analyzer.Analyze(dirAsts)
}

// Analyze 解析并分析单个文件或目录；诊断输出到 stderr，有错误时返回 ErrReported。
func Analyze(path string) (*globals.Package, error) {
	// 绝对化入口路径：包路径参与符号名哈希与缓存判定，不能随启动目录/给参形式漂移
	path = stlerr.MustWith(filepath.Abs(path))
	info, err := stlerr.ErrorWith(os.Stat(path))
	if err != nil {
		return nil, err
	}
	reporter := report.NewReporter()
	var pkg *globals.Package
	if info.IsDir() {
		pkg, err = analyzeDir(path, reporter)
	} else {
		pkg, err = analyzeFile(path, reporter)
	}
	if reporter.HasErrors() {
		reporter.Print(os.Stderr)
		return nil, report.ErrReported
	} else if err != nil {
		return nil, err
	}
	if reporter.HasWarns() {
		reporter.Print(os.Stderr)
	}
	return pkg, nil
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

// Analyze 分析程序；有诊断错误时返回 ErrReported。
func (a *Analyzer) Analyze(program *ast.File) (*globals.Package, error) {
	if err := a.analyzePackageImport(program); err != nil {
		return nil, err
	}
	a.analyzeGlobalType(program)
	a.analyzeGlobalValue(program)
	if a.reporter.HasErrors() {
		return nil, report.ErrReported
	}
	return a.ir, nil
}

func (a *Analyzer) analyzePackageImport(program *ast.File) error {
	err := a.importBuildin()
	if err != nil {
		return err
	}
	for _, g := range program.Globals {
		importAst, ok := g.(*ast.Import)
		if !ok {
			continue
		}
		err = a.analyzeImport(importAst)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) analyzeGlobalType(program *ast.File) {
	for _, g := range program.Globals {
		func() {
			defer a.recoverFromAbort()
			decl := a.analyzeTypePreDecl(g)
			if decl == nil {
				return
			}
			a.ir.Globals = append(a.ir.Globals, decl)
		}()
	}
	for _, g := range program.Globals {
		func() {
			defer a.recoverFromAbort()
			a.analyzeTypeDecl(g)
		}()
	}
	for _, g := range program.Globals {
		func() {
			defer a.recoverFromAbort()
			a.analyzeTypeDef(g)
		}()
	}
	// 检查循环引用
	for _, g := range program.Globals {
		t, ok := g.(*ast.TypeDef)
		if !ok {
			continue
		}
		func() {
			defer a.recoverFromAbort()
			ct, ok := a.scope.LookupType(t.Name.OriginText)
			if !ok {
				return
			}
			if types.CheckRecursion(ct) {
				a.reporter.Errorf(
					t.Name.Position,
					report.Errors.CircularReference,
				)
			}
		}()
	}
}

func (a *Analyzer) analyzeGlobalValue(program *ast.File) {
	for _, v := range program.Globals {
		func() {
			defer a.recoverFromAbort()
			a.analyzeGlobalValueDecl(v)
		}()
	}

	for _, v := range program.Globals {
		func() {
			defer a.recoverFromAbort()
			def := a.analyzeGlobalValueDef(v)
			if def == nil {
				return
			}
			a.ir.Globals = append(a.ir.Globals, def)
		}()
	}
}

// analyzeAbort 当前声明/函数分析失败（诊断已记录），由恢复点隔离后继续分析其他声明。
type analyzeAbort struct{}

func (a *Analyzer) abort() {
	panic(analyzeAbort{})
}

// recoverFromAbort 隔离当前声明/函数分析中的中止；其他 panic 继续上抛为 ICE。
func (a *Analyzer) recoverFromAbort() {
	if r := recover(); r != nil {
		if _, aborted := r.(analyzeAbort); !aborted {
			panic(r)
		}
	}
}

// errorf 记录诊断，不中断分析；需要中断当前声明/函数时调用 abort。
func (a *Analyzer) errorf(pos reader.Position, err report.ErrorType, args ...any) {
	a.reporter.Errorf(pos, err, args...)
}

// isInvalidType 是否为错误恢复用的占位类型。
func isInvalidType(t hir.Type) bool {
	_, ok := t.(types.InvalidType)
	return ok
}

// isUnitType 是否为 unit 类型。
func isUnitType(t hir.Type) bool {
	return stlval.Is[types.UnitType](t)
}
