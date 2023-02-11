package parse

import (
	"bytes"
	"os"
	"sort"

	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/stl/list"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/set"
)

// Package 包
type Package struct {
	Priority   uint                           // 优先级，优先级越高越应该优先处理
	Path       stlos.Path                     // 包路径
	Globals    *list.SingleLinkedList[Global] // 全局语句
	importMap  map[string]*Package            // 导入的包
	includeMap *set.LinkedHashSet[*Package]   // 包含的包，最后面的包应该最先查找
}

// NewPackage 新建包
func NewPackage(path stlos.Path) *Package {
	return &Package{
		Priority:   0,
		Path:       path,
		Globals:    list.NewSingleLinkedList[Global](),
		importMap:  make(map[string]*Package),
		includeMap: set.NewLinkedHashSet[*Package](),
	}
}

// 词法-语法分析单文件
func parseFile(pkg *Package, path stlos.Path) error {
	content, err := os.ReadFile(string(path))
	if err != nil {
		return err
	}
	reader := bytes.NewReader(content)
	lexer := lex.NewLexer(path, reader)
	parser := newParser(pkg, lexer)
	return parser.parse()
}

// 词法-语法分析包
func parsePackage(path stlos.Path) (*Package, error) {
	pkg := NewPackage(path)
	pkgs[path] = nil

	filePaths, err := os.ReadDir(string(path))
	if err != nil {
		return nil, err
	}
	for _, fp := range filePaths {
		if fp.IsDir() {
			continue
		}
		fp := path.Join(stlos.Path(fp.Name()))
		if fp.GetExtension() != "sim" {
			continue
		}

		err := parseFile(pkg, fp)
		if err != nil {
			return nil, err
		}
	}

	pkgs[path] = pkg
	return pkg, nil
}

// Parse 词法-语法分析
func Parse(path stlos.Path) ([]*Package, error) {
	// 获取绝对路径
	path, err := path.GetAbsolute()
	if err != nil {
		panic(err)
	}

	// 对路径进行词法-语法分析
	var mainPkg *Package
	if !path.IsDir() {
		mainPkg = NewPackage(path)
		pkgs[path] = nil
		if err := parseFile(mainPkg, path); err != nil {
			return nil, err
		}
		pkgs[path] = mainPkg
	} else {
		var err error
		mainPkg, err = parsePackage(path)
		if err != nil {
			return nil, err
		}
	}

	// 计算优先级
	var walk func(p *Package)
	walk = func(p *Package) {
		for _, dst := range p.importMap {
			if p.Priority+1 > dst.Priority {
				dst.Priority = p.Priority + 1
				walk(dst)
			}
		}
	}
	walk(mainPkg)

	// 优先级排序，优先级越高，就处于依赖树的越上层
	pkgsCpy := make([]*Package, len(pkgs))
	var i int
	for _, p := range pkgs {
		pkgsCpy[i] = p
		i++
	}
	sort.Slice(
		pkgsCpy, func(i, j int) bool {
			return pkgsCpy[i].Priority > pkgsCpy[j].Priority
		},
	)

	return pkgsCpy, nil
}
