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

	globals *list.SingleLinkedList[hir.Global] // 全局
}

// NewAnalyser 新建语义分析器
func NewAnalyser() *Analyser {
	return &Analyser{
		symbols: make(map[*parse.Package]*symbolTable),
		globals: list.NewSingleLinkedList[hir.Global](),
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

	// 类型定义
	if err := self.analysePackageTypedef(*ast); err != nil {
		return nil, err
	}
	// 对象声明
	declAsts, err := self.analysePackageObjectDecl(*ast)
	if err != nil {
		return nil, err
	}
	// 对象定义
	if err := self.analysePackageObjectDef(declAsts); err != nil {
		return nil, err
	}
	// 特征实现检查
	if err := self.analysePackageImplTraitCheck(); err != nil {
		return nil, err
	}
	return self.globals, nil
}

// 包类型定义
func (self *Analyser) analysePackageTypedef(astPkg parse.Package) utils.Error {
	// 声明
	typeAsts := list.NewSingleLinkedList[parse.TypeDef]()
	traitAsts := list.NewSingleLinkedList[parse.Trait]()
	var errs []utils.Error
	for iter := astPkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		switch ast := iter.Value().(type) {
		case *parse.TypeDef:
			if len(ast.GenericParams) == 0 {
				if !self.symbol.declType(
					ast.Public,
					hir.NewTypedef(hir.NewPkgPath(astPkg.Path), ast.Name.Source, hir.NewTypeNone()),
				) {
					errs = append(errs, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration))
				} else {
					typeAsts.PushBack(*ast)
				}
			} else {
				if err := self.analyseGenericTypeDecl(*ast); err != nil {
					errs = append(errs, err)
				}
			}
		case *parse.Trait:
			if !self.symbol.declTrait(
				ast.Public,
				hir.NewTrait(hir.NewPkgPath(astPkg.Path), ast.Name.Source, nil),
			) {
				errs = append(errs, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration))
			} else {
				traitAsts.PushBack(*ast)
			}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}
	// 类型定义
	for iter := typeAsts.Iterator(); iter.HasValue(); iter.Next() {
		def, err := self.analyseTypedef(iter.Value())
		if err != nil {
			errs = append(errs, err)
		} else {
			self.globals.PushBack(def)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 循环检测
	for iter := typeAsts.Iterator(); iter.HasValue(); iter.Next() {
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
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}

	// 特征定义
	for iter := traitAsts.Iterator(); iter.HasValue(); iter.Next() {
		def, err := self.analyseTrait(iter.Value())
		if err != nil {
			errs = append(errs, err)
		} else {
			self.globals.PushBack(def)
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}
	return nil
}

// 包对象声明
func (self *Analyser) analysePackageObjectDecl(astPkg parse.Package) (
	*list.SingleLinkedList[parse.Global], utils.Error,
) {
	decls := list.NewSingleLinkedList[parse.Global]()

	var errs []utils.Error
	for iter := astPkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		var skip bool
		var err utils.Error
		switch ast := iter.Value().(type) {
		case *parse.GlobalValue:
			_, err = self.analyseGlobalValueDecl(*ast)
		case *parse.Function:
			if len(ast.GenericParams) == 0 {
				_, err = self.analyseFunctionDecl(*ast)
			} else {
				err = self.analyseGenericFunctionDecl(*ast)
				skip = true
			}
		case *parse.Method:
			if len(ast.SelfGenericParams) == 0 && len(ast.GenericParams) == 0 {
				_, err = self.analyseMethodDecl(*ast)
			} else {
				err = self.analyseGenericMethodDecl(*ast)
				skip = true
			}
		default:
			continue
		}
		if err != nil {
			errs = append(errs, err)
		} else if !skip {
			decls.PushBack(iter.Value())
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	return decls, nil
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

// 包特征实现检查
func (self *Analyser) analysePackageImplTraitCheck() utils.Error {
	var errs []utils.Error
	for iter := self.globals.Iterator(); iter.HasValue(); iter.Next() {
		if td, ok := iter.Value().(*hir.Typedef); ok {
			for iter := td.Impls.Iterator(); iter != nil; iter.Next() {
				trait := iter.Value()
				if !td.IsImpl(trait) {
					errs = append(errs, utils.Errorf(utils.Position{}, "not impl %s", trait.Name)) // TODO: Position
				}
				if !iter.HasNext() {
					break
				}
			}
		}
	}
	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return utils.NewMultiError(errs...)
	}
	return nil
}
