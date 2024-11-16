package analyse

import (
	"fmt"
	"os"

	"github.com/kkkunny/stl/container/tuple"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/parse"
)

type importPackageError interface {
	error
	importPackage()
}

type importPackageDuplicationError struct {
	name string
}

func (err *importPackageDuplicationError) Error() string {
	return fmt.Sprintf("duplicate package name `%s`", err.name)
}

func (err *importPackageDuplicationError) importPackage() {}

type importPackageCircularError struct {
	chain []*hir.Package
}

func (err *importPackageCircularError) Error() string {
	return "circular import package"
}

func (err *importPackageCircularError) importPackage() {}

type importPackageInvalidError struct {
	path stlos.FilePath
}

func (err *importPackageInvalidError) Error() string {
	return "invalid package"
}

func (err *importPackageInvalidError) importPackage() {}

// importPackage 导入包
func (self *Analyser) importPackage(pkgPath stlos.FilePath, name string, importAll bool) (dstPkg *hir.Package, err importPackageError) {
	name = stlval.Ternary(name != "", name, pkgPath.Base())

	defer func() {
		if err != nil {
			return
		}
		if importAll {
			self.pkg.AddLinkedPackage(dstPkg)
		} else {
			self.pkg.SetExternPackage(name, dstPkg)
		}
	}()

	// 检查包名冲突
	if !importAll && tuple.Pack2(self.pkg.GetExternPackage(name)).E2() {
		return nil, &importPackageDuplicationError{name: name}
	}

	// 检查包地址
	pathInfo, originErr := os.Stat(string(pkgPath))
	if originErr != nil || !pathInfo.IsDir() {
		return nil, &importPackageInvalidError{path: pkgPath}
	}

	dstPkg = hir.NewPackage(pkgPath)

	// 检查循环依赖
	pkgChan := make([]*hir.Package, 0, self.importStack.Length())
	for iter := self.importStack.Iterator(); iter.Next(); {
		pkg := iter.Value()
		pkgChan = append(pkgChan, pkg)
		if pkg.Equal(dstPkg) {
			return nil, &importPackageCircularError{chain: pkgChan}
		}
	}

	// 有缓存的直接返回
	if self.allPkgs.Contain(dstPkg.Path()) {
		return self.allPkgs.Get(dstPkg.Path()), nil
	}

	// 分析包
	self.allPkgs.Set(dstPkg.Path(), dstPkg)
	self.importStack.Push(self.pkg)
	defer func() {
		self.importStack.Pop()
	}()
	empty, originErr := analyseSonPackage(self, dstPkg)
	if originErr != nil || empty {
		return nil, &importPackageInvalidError{path: pkgPath}
	}
	return dstPkg, nil
}

// Analyse 语义分析
func Analyse(path stlos.FilePath) (*hir.Package, error) {
	asts, err := parse.Parse(path)
	if err != nil {
		return nil, err
	}
	return New(path).Analyse(asts), nil
}

// 语义分析子包
func analyseSonPackage(parent *Analyser, pkg *hir.Package) (bool, error) {
	asts, err := parse.Parse(pkg.Path())
	if err != nil {
		return false, err
	} else if asts.Empty() {
		return true, nil
	}
	newSon(parent, pkg).Analyse(asts)
	return false, nil
}
