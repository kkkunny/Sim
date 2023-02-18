package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/list"
	"github.com/kkkunny/stl/set"
)

// Analyser 语义分析器
type Analyser struct {
	symbol  *symbolTable                    // 当前符号表
	symbols map[*parse.Package]*symbolTable // 符号表

	typedefTemp map[string]parse.TypeDef // 类型定义缓存
}

// NewAnalyser 新建语义分析器
func NewAnalyser() *Analyser {
	return &Analyser{
		symbols:     make(map[*parse.Package]*symbolTable),
		typedefTemp: make(map[string]parse.TypeDef),
	}
}

// Analyse 语义分析
func (self *Analyser) Analyse(asts []*parse.Package) (*hir.Package, utils.Error) {
	hirs := list.NewSingleLinkedList[hir.Global]()
	for _, ast := range asts {
		pkgHirs, err := self.analysePackage(ast)
		if err != nil {
			return nil, err
		}
		hirs.Contact(pkgHirs)
	}
	return &hir.Package{Globals: hirs}, nil
}

// 包
func (self *Analyser) analysePackage(ast *parse.Package) (*list.SingleLinkedList[hir.Global], utils.Error) {
	self.symbol = newPkgSymbolTable(ast.Path)
	self.symbols[ast] = self.symbol

	globals := list.NewSingleLinkedList[hir.Global]()

	// 类型定义
	typedefs, err := self.analysePackageTypedef(*ast)
	if err != nil {
		return nil, err
	}
	for iter := typedefs.Iterator(); iter.HasValue(); iter.Next() {
		globals.PushBack(iter.Value())
	}
	// 对象声明
	objects, declAsts, err := self.analysePackageObjectDecl(*ast)
	if err != nil {
		return nil, err
	}
	globals.Contact(objects)
	// 对象定义
	if err = self.analysePackageObjectDef(declAsts); err != nil {
		return nil, err
	}
	return globals, nil
}

// 包类型定义
func (self *Analyser) analysePackageTypedef(astPkg parse.Package) (*list.SingleLinkedList[*hir.Typedef], utils.Error) {
	defAsts := list.NewSingleLinkedList[parse.TypeDef]()
	// 声明
	for iter := astPkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		if ast, ok := iter.Value().(*parse.TypeDef); ok {
			if !self.symbol.declType(
				ast.Public,
				hir.NewTypedef(hir.NewPkgPath(astPkg.Path), ast.Name.Source, hir.NewTypeNone()),
			) {
				return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
			}
			defAsts.PushBack(*ast)
		}
	}
	// 定义
	defs := list.NewSingleLinkedList[*hir.Typedef]()
	var errs []utils.Error
	for iter := defAsts.Iterator(); iter.HasValue(); iter.Next() {
		def, err := self.analyseTypedef(iter.Value())
		if err != nil {
			errs = append(errs, err)
		} else {
			defs.PushBack(def)
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	if len(self.typedefTemp) != 0 {
		panic("unreachable")
	}

	// 循环检测
	for iter := defAsts.Iterator(); iter.HasValue(); iter.Next() {
		astName := iter.Value().Name
		def, ok := self.symbol.lookupType(astName.Source)
		if !ok || def.data.Target.IsNone() {
			panic("unreachable")
		}
		if self.checkTypeCircle(set.NewLinkedHashSet[*hir.Typedef](), hir.NewTypeTypedef(def.data)) {
			errs = append(errs, utils.Errorf(astName.Pos, errCircularReference))
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}
	return defs, nil
}

// 包对象声明
func (self *Analyser) analysePackageObjectDecl(astPkg parse.Package) (
	*list.SingleLinkedList[hir.Global], *list.SingleLinkedList[parse.Global], utils.Error,
) {
	globals := list.NewSingleLinkedList[hir.Global]()
	decls := list.NewSingleLinkedList[parse.Global]()

	var errs []utils.Error
	for iter := astPkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		var global hir.Global
		var err utils.Error
		switch ast := iter.Value().(type) {
		case *parse.GlobalValue:
			global, err = self.analyseGlobalValueDecl(*ast)
		case *parse.Function:
			global, err = self.analyseFunctionDecl(*ast)
		case *parse.Method:
			global, err = self.analyseMethodDecl(*ast)
		default:
			continue
		}
		if err != nil {
			errs = append(errs, err)
		} else {
			globals.PushBack(global)
			decls.PushBack(iter.Value())
		}
	}
	if len(errs) == 1 {
		return nil, nil, errs[0]
	} else if len(errs) > 1 {
		return nil, nil, utils.NewMultiError(errs...)
	}

	return globals, decls, nil
}

// 包对象定义
func (self *Analyser) analysePackageObjectDef(asts *list.SingleLinkedList[parse.Global]) utils.Error {
	var errs []utils.Error
	for iter := asts.Iterator(); iter.HasValue(); iter.Next() {
		var err utils.Error
		switch ast := iter.Value().(type) {
		case *parse.GlobalValue:
			err = self.analyseGlobalValueDef(*ast)
		case *parse.Function:
			err = self.analyseFunctionDef(*ast)
		case *parse.Method:
			err = self.analyseMethodDef(*ast)
		default:
			panic("unreachable")
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	return nil
}
